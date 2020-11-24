![Build](https://github.com/Geo25rey/ipmail/workflows/Build/badge.svg)
![CodeQL](https://github.com/Geo25rey/ipmail/workflows/CodeQL/badge.svg)

# InterPlanetary Mail
InterPlanetary Mail (IPMail) is a decentralized email alternative, which uses IPFS to send and receive encrypted messages. 

Download the latest build [here](https://github.com/Geo25rey/ipmail/actions?query=is%3Asuccess+branch%3Amaster+workflow%3ABuild). Note that to download these builds you must be logged in to your GitHub account.

## Building
1) Install the go compiler [here](https://golang.org/dl/) or from your favorite package manager.
2) Run `./build.sh`<br/>
   <b>OR</b><br/>
   Run the following commands:
   ```bash
   mkdir build && cd build
   go build ..
   ```
3) Building done. Look in the build build folder for the `ipmail` program

## Basic Use
1) Run `build/ipmail`
2) Wait a few seconds for things to load
3) Follow the setup prompts
4) Run the `help` command to see a list of all commands
