import os,io,sys
from flask import * 
import struct

app = Flask(__name__,static_folder="./static")

@app.route("/",methods=["GET"])
def MainPage():
    return render_template("index.html")

@app.route("/about",methods=["GET"])
def About():
    return render_template("about.html")
    
app.run("0.0.0.0",80,debug=True)
