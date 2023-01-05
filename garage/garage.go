package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type RPC_Request struct {
	JSONver string   `json:"jsonrpc"`
	ID      int      `json:"id"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
}

type ErrorObj struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}
type RPC_Response struct {
	JSONver string   `json:"jsonrpc"`
	ID      int      `json:"id"`
	Result  int      `json:"result"`
	Error   ErrorObj `json:"error"`
	Data    string   `json:"data"`
}

//login check constants
const NOT_FOUND = 0x23
const OK = 0x44
const WRONG_CREDS = 0x33

const ERROR = 0x55           //default error
const SYS_GRADE_ERROR = 0x21 // system (server) got errored.
const USER_GRADE_ERROR = 0x22

const GARAGE_PORT = "3243"

/*
	We are going to have couple of Garage-server RPC methods.
	List:
		grg_createLogin, params: [username,address,password] // will have to create tron acc in web3js
		grg_checkLogin, params: [username,password]
		grg_getCryptoAddress, params: [username]
		grg_changePassword , params: [username,oldpasswd,newpasswd]
		grg_deleteAccount, params: [username,password]
		grg_createChatRoom, params: [username,number of users,myPublicKey] , returns ID of chat room
		grg_getChatMembers, params: [id_of_group]
		grg_addChatMember ,params: [username,password,memberaddress,id]
		grg_getChatPublicKey ,params: [username,password,id,address]
		grg_addPublicKey ,params: [username,passwd,pk,id_group]


	Example:
		{"jsonrpc":"2.0","id":1,"method":"grg_createLogin","params":["gavrilo","TGR8WXbGPBqXrGKB5z5oAmPf2wyv1y3Htq","gasha2015"]}

*/
func send_stat_rpc(res http.ResponseWriter, msg string, typed string, code int) {

	var resp RPC_Response
	resp.JSONver = "2.0"
	resp.ID = 1
	resp.Result = code
	if typed == "e" {
		resp.Error = ErrorObj{
			Code:    code,
			Message: msg,
			Data:    msg,
		}
		resp.Result = ERROR
	} else {
		resp.Data = msg
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	json.NewEncoder(res).Encode(resp)

}
func encrypt_hash(passwd string) string {
	hash := sha256.New()
	hash.Write([]byte(passwd))
	hashed := hash.Sum(nil)
	res := fmt.Sprintf("%x", hashed)
	return res
}
func createNewLogin(params []string) int {

	files, err := os.Stat("./garaged_data")
	if err != nil {
		return 0
	}
	if files.Mode()&(1<<2) == 0 {
		return 0
	}

	encrypted := encrypt_hash(params[2])
	if encrypted == "e" {
		return 0
	}
	bytedData := []byte(params[1] + ":" + encrypted)
	err = ioutil.WriteFile("./garaged_data/"+params[0], bytedData, 0777)
	if err != nil {
		fmt.Println("error writing to file.")
		return 0
	}
	return 1
}

/* This function checks if the provided user already exists in the db. if does, returns true, if not ,returns false. */
func checkExistingUser(user string, address string) bool {

	files, err := ioutil.ReadDir("./garaged_data")
	if err != nil {
		err = os.Mkdir("./garaged_data", 0777)
		if err != nil {
			log.Fatal("Could not create garaged_data folder. Insufficient permissions.")
		}
	}

	for i := 0; i < len(files); i++ {
		if files[i].Name() == user {
			return true
		}
	}
	return false

}
func getCryptoAddress(username string) string {
	data, err := ioutil.ReadFile("./garaged_data/" + username)
	if err != nil {
		return "NOT_FOUND"
	}
	vard := string(data)
	dat := strings.Split(vard, ":")
	return dat[0]

}
func deleteAccount(params []string) int {
	err := os.Remove("./garaged_data/" + params[0])
	if err != nil {
		return NOT_FOUND
	} else {
		return OK
	}
}

func LoginUserCheck(params []string) int {

	usr := params[0]
	passwd := params[1]
	data, err := ioutil.ReadFile("./garaged_data/" + usr)
	if err != nil {
		return NOT_FOUND
	}
	dt := string(data)
	dtd := strings.Split(dt, ":")
	if dtd[1] == encrypt_hash(passwd) {
		return OK
	} else {
		return WRONG_CREDS
	}

}

func changePassword(params []string) int {

	logined := LoginUserCheck(params)
	if logined != OK {
		return logined
	}
	data := encrypt_hash(params[2])
	addr := getCryptoAddress(params[0])
	data = addr + ":" + data

	fmt.Printf("[ LOG ] password changed %s\n", data)

	err := ioutil.WriteFile("./garaged_data/"+params[0], []byte(data), 0777)
	if err != nil {
		return SYS_GRADE_ERROR
	}
	return OK

}
func createChatGroup(params []string) int {

	id_group := 0
	groups, err := ioutil.ReadDir("./groups")
	if err != nil {
		return ERROR
	}
	if len(groups) != 0 {
		id_group = len(groups)
	}
	data := ""

	data = "admin:" + getCryptoAddress(params[0]) + "\n" + getCryptoAddress(params[0])

	err = ioutil.WriteFile("./groups/"+strconv.Itoa(id_group)+"_"+string(params[1]), []byte(data), 0777)
	if err != nil {
		return ERROR
	}
	datd := getCryptoAddress(params[0])
	if datd == "NOT_FOUND" {
		return ERROR
	}
	datd += "," + params[2]
	err = ioutil.WriteFile("./keys/"+strconv.Itoa(id_group), []byte(datd), 0777)
	if err != nil {
		return ERROR
	}

	return id_group

}
func getPublicKey(params []string) string {

	addr := getCryptoAddress(params[0])
	if addr == "NOT_FOUND" {
		return "PERM_DENIED"
	}
	id := params[2]
	dir, err := ioutil.ReadDir("./groups")
	found := false
	addr_in_group := false
	if err != nil {
		return "PERM_DENIED"
	}
	for i := 0; i < len(dir); i++ {
		name := string(dir[i].Name())
		dat := strings.Split(name, "_")
		if dat[0] == id {
			id = name
			found = true
		}
	}
	if found {
		file, _ := ioutil.ReadFile("./groups/" + id)
		membs := strings.Split(string(file), "\n")
		for i := 1; i < len(membs); i++ {
			if membs[i] == addr {
				addr_in_group = true
			}
		}
		if addr_in_group {
			d, err := ioutil.ReadFile("./keys/" + params[2])
			if err != nil {
				return "NOT_FOUND"
			}
			keys := strings.Split(string(d), "\n")
			for i := 0; i < len(keys); i++ {
				keyd := strings.Split(keys[i], ",")

				if keyd[0] == params[3] {
					return keyd[1]
				}
			}
		} else {
			return "PERM_DENIED"
		}
	}
	return "PERM_DENIED"

}
func getChatMembers(id string) string {
	dir, err := ioutil.ReadDir("./groups")
	found := false
	if err != nil {
		return "NOT_FOUND"
	}
	for i := 0; i < len(dir); i++ {
		name := string(dir[i].Name())
		dat := strings.Split(name, "_")

		if dat[0] == id {
			id = name
			found = true
		}
	}

	if found == true {
		file, err := ioutil.ReadFile("./groups/" + id)
		if err != nil {
			return "NOT_FOUND"
		}
		fild := strings.Split(string(file), "\n")
		res := ""
		for i := 1; i < len(fild); i++ {
			res += string(fild[i]) + "\n"
		}
		file = []byte(strings.Replace(string(res), "\n", ",", len(res)))
		return string(file)
	}

	return "NOT_FOUND"
}

/* This function has to first check the admin of the group, and if you are admin, good to go, just append to file. */
func addChatMember(params []string) int {
	addr := getCryptoAddress(params[0])
	if addr == "NOT_FOUND" {
		return NOT_FOUND

	}
	id := params[3]
	dir, err := ioutil.ReadDir("./groups")
	found := false
	if err != nil {
		return SYS_GRADE_ERROR
	}
	for i := 0; i < len(dir); i++ {
		name := string(dir[i].Name())
		dat := strings.Split(name, "_")
		if dat[0] == id {
			id = name
			found = true
		}
	}
	if found {
		d, e := ioutil.ReadFile("./groups/" + id)
		if e != nil {
			return SYS_GRADE_ERROR
		} else {
			data := string(d)
			members := strings.Split(string(d), "\n")
			admin := strings.Split(members[0], ":")[1]
			if admin != addr {
				return ERROR
			}
			data += "\n" + params[2]
			err := ioutil.WriteFile("./groups/"+id, []byte(data), 0777)
			if err != nil {
				return SYS_GRADE_ERROR
			}
			return OK
		}
	}
	return NOT_FOUND

}

/* grg_addPublicKey ,params: [username,passwd,pk,id_group] */
func addPublicKey(params []string) int {

	logd := LoginUserCheck(params)
	if logd != OK {
		return ERROR
	}
	fd, err := ioutil.ReadFile("./keys/" + params[3])

	if err != nil {
		return NOT_FOUND
	}
	addr := getCryptoAddress(params[0])
	if addr == "NOT_FOUND" {
		return ERROR
	}
	data := string(fd) + "\n" + addr + "," + params[2]
	fmt.Println(data)
	err = ioutil.WriteFile("./keys/"+params[3], []byte(data), 0777)
	if err != nil {
		return ERROR
	}
	return OK
}
func start_daemon(certPath string, keyPath string) {

	s := &http.Server{
		Addr:    ":" + GARAGE_PORT,
		Handler: nil,
	}
	http.HandleFunc("/db", func(res http.ResponseWriter, req *http.Request) {

		if req.Method == "POST" {

			body, err := ioutil.ReadAll(req.Body)
			fmt.Printf("[ LOG ] Received request:%s body: %s\n", req.Method, body)
			if err != nil {

				send_stat_rpc(res, "error reading body", "e", SYS_GRADE_ERROR)

			}
			rawIn := json.RawMessage(body)
			bytes, err := rawIn.MarshalJSON()
			if err != nil {
				log.Fatal("marshalJSON() failed.")
			}
			var req RPC_Request
			fmt.Println(req.Method)
			err = json.Unmarshal([]byte(bytes), &req)

			if err != nil {
				log.Fatal("unmarshal err.")
			}
			if req.JSONver != "2.0" {
				send_stat_rpc(res, "invalid json version", "e", USER_GRADE_ERROR)
			}

			if req.Method == "grg_getChatPublicKeys" {

				logd := LoginUserCheck(req.Params)
				if logd != OK {
					send_stat_rpc(res, "invalid creds", "e", USER_GRADE_ERROR)
				} else {
					pkeys := getPublicKey(req.Params)
					if pkeys == "PERM_DENIED" {
						send_stat_rpc(res, "permission denied", "e", USER_GRADE_ERROR)
					}
					if pkeys == "NO_KEYS" {
						send_stat_rpc(res, "no public keys found", "e", USER_GRADE_ERROR)
					} else {
						send_stat_rpc(res, pkeys, "s", OK)
					}
				}

			}

			if req.Method == "grg_getChatMembers" {

				members := getChatMembers(req.Params[0])
				if members == "NOT_FOUND" {
					send_stat_rpc(res, "group not found", "e", USER_GRADE_ERROR)
				} else {
					send_stat_rpc(res, members, "s", OK)

				}
			}
			if req.Method == "grg_addPublicKey" {

				added := addPublicKey(req.Params)
				if added != OK {
					send_stat_rpc(res, "err", "e", added)
				} else {
					send_stat_rpc(res, "OK", "s", OK)

				}
			}
			if req.Method == "grg_addChatMember" {
				logd := LoginUserCheck(req.Params)
				if logd != OK {
					send_stat_rpc(res, "invalid credidentials", "e", USER_GRADE_ERROR)

				} else {
					added := addChatMember(req.Params)
					if added != OK {
						send_stat_rpc(res, "access denied", "e", USER_GRADE_ERROR)
					} else {
						send_stat_rpc(res, "OK", "s", OK)
					}
				}
			}
			if req.Method == "grg_checkLogin" {

				logined := LoginUserCheck(req.Params)
				if logined == OK {
					send_stat_rpc(res, "OK", "s", OK)

				} else if logined == NOT_FOUND {
					send_stat_rpc(res, "user not found", "e", USER_GRADE_ERROR)

				} else {
					send_stat_rpc(res, "invalid credidentials", "e", USER_GRADE_ERROR)
				}
			}
			if req.Method == "grg_getCryptoAddress" {
				addr := getCryptoAddress(req.Params[0])
				if addr == "NOT_FOUND" {
					send_stat_rpc(res, "could not find address", "e", USER_GRADE_ERROR)
				} else {
					send_stat_rpc(res, addr, "s", OK)
				}
			}

			if req.Method == "grg_createChatRoom" {

				gid := createChatGroup(req.Params)
				if gid == ERROR {
					send_stat_rpc(res, "error", "e", gid)

				} else {
					send_stat_rpc(res, "room_id:"+strconv.Itoa(gid), "s", gid)
				}

			}

			if req.Method == "grg_changePassword" {

				changed := changePassword(req.Params)
				if changed == OK {
					send_stat_rpc(res, "OK", "s", OK)

				} else {
					send_stat_rpc(res, "Error", "e", changed)
				}
			}

			if req.Method == "grg_createLogin" {
				exists := checkExistingUser(req.Params[0], req.Params[1])
				if !exists {
					stat := createNewLogin(req.Params)
					if stat != 1 {
						send_stat_rpc(res, "error creating new login.", "e", SYS_GRADE_ERROR)
					} else {
						send_stat_rpc(res, "OK", "s", OK)
					}

				} else {
					send_stat_rpc(res, "account already exists.", "e", USER_GRADE_ERROR)
				}
			}

			if req.Method == "grg_deleteAccount" {
				chk := LoginUserCheck(req.Params)
				if chk == OK {
					del := deleteAccount(req.Params)
					if del == OK {
						send_stat_rpc(res, "OK", "s", OK)
					} else {
						send_stat_rpc(res, "account_not_found", "e", USER_GRADE_ERROR)

					}
				} else {
					send_stat_rpc(res, "invalid credidentials", "e", USER_GRADE_ERROR)
				}
			}

		} else {
			fmt.Fprint(res, "invalid request got:"+req.Method)
		}
	})
	fmt.Println("~ Garage server running...")
	s.ListenAndServeTLS(certPath, keyPath)

}

func main() {

	if len(os.Args) < 3 {
		log.Fatal("\n************ Garage Server - backend for Borderline ************\n\tProvide path to private key, and certificate.")
	}
	key := os.Args[1]
	cert := os.Args[2]
	start_daemon(cert, key)
}
