#!/usr/bin/python2.7

from flask import Flask, request
import os
import requests

app = Flask(__name__)

@app.route("/")
def client():
    return 'Hi %s, this is %s\n' % (
        request.remote_addr,
        os.popen('hostname -i').read().strip())

if __name__=='__main__':
    app.run('0.0.0.0', 8080)
