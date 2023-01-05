import os,io,sys
import struct 
import tronpy
import base58
from Crypto.Cipher import PKCS1_OAEP
import random
from Crypto.PublicKey import RSA
import ecdsa
from multiprocessing import Process
import json
import requests
from socket import * 



"""

    How does this API send a message?

    Using RSA, we encrypt a message (using pubkey)
    We need a struct to know where we are sending messages.
    If we are sending direct messages, our struct looks like this:
    usr:{address}:RSAencrypted
    and that is being sent directly to user
    but when your daemon receives messages. if he sees 


"""
#Types of messages on Nemesis Network
DIRECT_MSG = 0x02
GROUP_MSG= 0x04
EXCHANGE = 0x06 #exchanging pub keys.
HEADER="NEMESISNT"
ERROR= 0x324

TRX_MSG_FEE = [1,20] #  for direct messaging , and SHOULD be for group messaging.
NEMESIS_WALLET ="TCeCDNKUubi6zH2fXT33wFH5JMKxwgpPWU" #our address (fee)
#NEMESIS_WALLET ="TGR8WXbGPBqXrGKB5z5oAmPf2wyv1y3Htq" #our company's address (fee)
GARAGE_URL = "https://localhost:3243/db" # URL for Garage Server (buying domain soon.)
INSUFFICIENT_FUNDS= 0x23

class Goblin(object):
    def __init__(self,pk):
        self.pk = pk         
        print(f"[ {self} ] Goblin instance.")
        self.cli = tronpy.Tron(network='mainnet')
        self.pk = tronpy.keys.PrivateKey(bytes.fromhex(self.pk))
        self.addr = self.cli.generate_address(self.pk)

    def sendToBoss(self,fromaddr,amount):
        transactionWithMsg = self.cli.trx.transfer(fromaddr,NEMESIS_WALLET,amount).build().sign(self.pk)
        ret = transactionWithMsg.broadcast().wait()
        return ret

    def broadcastMessage(self,toUser,message,type):
        fee = TRX_MSG_FEE
        if(type != DIRECT_MSG):    
            fee=[1,8]
    
        sender = self.addr["base58check_address"]
        if(self.cli.get_account_balance(sender) < 3):
            return INSUFFICIENT_FUNDS
        ret = self.sendToBoss(sender,fee[1])

        transactionWithMsg = self.cli.trx.transfer(sender,toUser,fee[0]).memo(message).build().sign(self.pk)
        ret = transactionWithMsg.broadcast().wait()
        return ret


    def generateKeyPair(self):
        return RSA.generate(2048)

    def createMessage(self,message,pubkey):       
        encryptor = PKCS1_OAEP.new(pubkey)
        msg =encryptor.encrypt(message.encode("utf-8"))
        msg = msg.hex()
        return msg

    def readLastMessage(self,addr):
        res = requests.get("https://api.trongrid.io/v1/accounts/"+addr+"/transactions")
        res = res.json()
        res = res['data']
  
        for trans in res:
            data = bytes.fromhex(trans['raw_data_hex'])
            if(HEADER.encode() in data):
                data= data.split(HEADER.encode()+b":")
                data=data[1]
                data=data.split(b"}")
                data=data[0].decode()+"}"
                return json.loads(data)   

        return None
        

    def sendDirectMessage(self,to,message,pubk):
        pubk = RSA.import_key(pubk)
        message =self.createMessage(message,pubk)
        msg = {
            "type":DIRECT_MSG,
            "id":self.addr["base58check_address"],
            "data":message
        }
        msg = HEADER+":"+json.dumps(msg)
        return self.broadcastMessage(to,msg,DIRECT_MSG)

    def sendGroupMessage(self,id,message,pubk,alreadyReceived):
        #fetching chat members from garage server...
        dat = {
            "jsonrpc":"2.0",
            "id":1,
            "method":"grg_getChatMembers",
            "params":[str(id)]
        }
        pubk = RSA.import_key(pubk)
        members = requests.post(GARAGE_URL,json=dat,verify="../garage/auth/localhost.crt")
        #checking server's response.
        if(len(members.text) > 0 ):
            members = members.json()['data'].split(",")
            if(len(members) > 0 ):
                message =self.createMessage(message,pubk)
                msg = {
                    "type":GROUP_MSG,
                    "id":id,
                    "received":alreadyReceived,
                    "data":message
                }
                msg = HEADER+":"+json.dumps(msg)
                memberz=[]
                for mem in members:
                    if(mem not in alreadyReceived and len(mem) >1):
                        memberz.append(mem)

                return self.broadcastMessage(random.choice(memberz),msg,GROUP_MSG)
            else:
                return ERROR
        else:
            return ERROR


        


if __name__ == '__main__':
    gob = Goblin("0DDA1FF5CC0D2A24CA462931F7272FF8A119DC0ACC08FA3562EC9009526AEDC3")
   #     gob.startDaemon()
    REC="TMtCrVeEThyHPo7vuuRvRXEZpbwLuAifzn" #scam receiver
    pubkey = open("./pub.pem","rb").read()
