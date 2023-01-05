<h1>Creating account</h1>

curl -X POST \
      -k -H 'Content-Type: application/json' \
     -d '{"jsonrpc":"2.0","id":1,"method":"grg_createLogin","params":["gavrilo","TGR8WXbGPBqXrGKB5z5oAmPf2wyv1y3Htq","password123"]}' \
     https://localhost:3243/db

<h1>Creating a chat room with maximum of 3 members</h1>

curl -X POST \
      -k -H 'Content-Type: application/json' \
     -d '{"jsonrpc":"2.0","id":1,"method":"grg_createChatRoom","params":["gavrilo","3"]}' \
     https://localhost:3243/db

<h1>Getting chat members...</h1>

curl -X POST \
      -k -H 'Content-Type: application/json' \
     -d '{"jsonrpc":"2.0","id":1,"method":"grg_getChatMembers","params":["0"]}' \
     https://localhost:3243/db
     
<h1>Getting public keys..</h1>

curl -X POST \
      -k -H 'Content-Type: application/json' \
     -d '{"jsonrpc":"2.0","id":1,"method":"grg_getChatPublicKeys","params":["gavrilo","password123","0"]}' \
     https://localhost:3243/db


<h1>Adding group chat member...</h1>

curl -X POST \
      -k -H 'Content-Type: application/json' \
     -d '{"jsonrpc":"2.0","id":1,"method":"grg_addGroupMember","params":["gavrilo","password123","0"]}' \
     https://localhost:3243/db
