# Nemesis
Nemesis Network is decentralized chatting application built entierly on TRON
<h1>How?</h1>
<p>This network uses transactions with "message" option to send encrypted data through TRON blockchain network.</p>
<p>Every message is encrypted with RSA and SHA-256 encryption methods using our Goblin API.</p>
<p>People share eachothers public keys,you have chat rooms, server id's, and thats how you exchange messages.</p>
<p><b>By the people, for the people.</b></p>

<h1>Backend</h1>
<p>our backend consists of a "garage" server written in Go.</p>
<p> it accepts JSON RPC requests that client sends to them.</p>
<h2><b>List of our JSON-RPC methods</b></h2>


    grg_createLogin, params: [username,address,password],
    grg_checkLogin, params: [username,password],
    grg_getCryptoAddress, params: [username],
    grg_changePassword , params: [username,oldpasswd,newpasswd],
    grg_deleteAccount, params: [username,password],
    grg_createChatRoom, params: [username,number of users,myPublicKey],
    grg_getChatMembers, params: [id_of_group],
    grg_addChatMember ,params: [username,password,memberaddress,id],
    grg_getChatPublicKey ,params: [username,password,id,address],
    grg_addPublicKey ,params: [username,passwd,pk,id_group],


<h1>Troubles</h1>
<p> - too pricey ($0.10 per message is A LOT) </p>
<p> - network that is hard to maintain ( need more people working on this. ) </p>
