# 0. Specify target network with defaults (dev env)
# 1. Create Identity
# 2. Create and Upload Keys
# 3. Spit out config.yaml
import argparse
from subprocess import call
import json, random, string, yaml, os

parser = argparse.ArgumentParser(description='Process some integers.')
parser.add_argument('--dev', action='store_true', help='Running in Dev mode')
parser.add_argument('--datadir', help='Target Data dir to Store Configs')
parser.add_argument('--apiport', default=8082 ,help='Node API Port')
parser.add_argument('--p2pport', default=38202, help='Node P2P Port')
parser.add_argument('--bootstraps', nargs='?', help='List of Bootsrap Peers')
args = parser.parse_args()
datadir = os.getcwd()+"/default"
ethnodeurl = "ws://127.0.0.1:9546"
ethaccountkey = '{"address":"89b0a86583c4444acfd71b463e0d3c55ae1412a5","crypto":{"cipher":"aes-128-ctr","ciphertext":"c779f8379d770d92cfc1ddd4a8f31d5a0adc8f2a0b2a1401370d3630f38c0c8a","cipherparams":{"iv":"36c168e73bf980fe75b0727f890a71ad"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"de1be16e3c981944d1eca2b8b27e4e6e0b5bfb43be0376a5d8889fa679a28122"},"mac":"cc128b815555ba1ead7cae9060e8842afca8356d06455bd9a5752ba6fcc092ef"},"id":"45e060a6-d2ae-43b8-922f-44829499d37d","version":3}'
ethaccountpwd = ''

def getDevContractAddresses():
    gopath = os.environ.get('GOPATH')
    scpath = gopath+"/src/github.com/centrifuge/centrifuge-ethereum-contracts/deployments/local.json"
    if os.path.exists(scpath):
        with open(scpath, 'r') as stream:
            fjson = json.load(stream)
            identityfactory = fjson['contracts']['IdentityFactory']['address']
            identityreg = fjson['contracts']['IdentityRegistry']['address']
            anchorrepository = fjson['contracts']['AnchorRepository']['address']
    else:
        return []

    return [identityfactory, identityreg, anchorrepository]

def createConfigFile():
    global datadir
    letters = string.ascii_lowercase
    rndname = ''.join(random.choice(letters) for i in range(5))
    if args.datadir:
        datadir = args.datadir
    if not os.path.exists(datadir):
        os.makedirs(datadir)
    stream = open('config.yaml.tpl', 'r')
    tpl = yaml.load(stream)
    tpl['ethereum']['nodeURL'] = ethnodeurl
    tpl['ethereum']['accounts']['main']['key'] = ethaccountkey
    tpl['ethereum']['accounts']['main']['password'] = ethaccountpwd
    tpl['keys']['p2p']['privateKey'] = tpl['keys']['p2p']['privateKey'].encode('utf-8').replace("${DATADIR}", datadir)
    tpl['keys']['p2p']['publicKey'] = tpl['keys']['p2p']['publicKey'].encode('utf-8').replace("${DATADIR}", datadir)
    tpl['keys']['ethauth']['privateKey'] = tpl['keys']['ethauth']['privateKey'].encode('utf-8').replace("${DATADIR}", datadir)
    tpl['keys']['ethauth']['publicKey'] = tpl['keys']['ethauth']['publicKey'].encode('utf-8').replace("${DATADIR}", datadir)
    tpl['keys']['signing']['privateKey'] = tpl['keys']['signing']['privateKey'].encode('utf-8').replace("${DATADIR}", datadir)
    tpl['keys']['signing']['publicKey'] = tpl['keys']['signing']['publicKey'].encode('utf-8').replace("${DATADIR}", datadir)
    tpl['storage']['path'] = tpl['storage']['path'].encode('utf-8').replace("${DATADIR}", datadir)
    tpl['nodePort'] = args.apiport
    tpl['p2p']['port'] = args.p2pport
    if args.dev:
        addrs = getDevContractAddresses()
        tpl['centrifugeNetwork'] = "testing"
        tpl['networks'] = {'testing': {'contractAddresses': {'IdentityFactory': addrs[0].encode('utf-8'), 'IdentityRegistry': addrs[1].encode('utf-8'), 'AnchorRepository': addrs[2].encode('utf-8')}}}
        if (args.bootstraps != None) and (len(args.bootstraps) > 0):
            tpl['networks']['testing']['bootstrapPeers'] = [args.bootstraps]
    fname = '/config_'+rndname+'.yaml'
    f = open(datadir+fname, "w+")
    with f as outfile:
        yaml.dump(tpl, outfile, default_flow_style=False)
    return datadir+fname

def createIdentity(configpath):
    call(['centrifuge', 'createidentity', '-c', configpath])
    genid = ""
    if os.path.exists('newidentity.json'):
        with open('newidentity.json', 'r') as stream:
            fjson = json.load(stream)
            genid = fjson['id']
    return genid

def addKeys(configpath):
    with open(configpath, 'r') as stream:
        tpl = yaml.load(stream)
        call(['centrifuge', 'generatekeys', '-t', 'secp256k1', '-p', tpl['keys']['ethauth']['privateKey'].encode('utf-8'), '-q', tpl['keys']['ethauth']['publicKey'].encode('utf-8'), '-c', configpath])
        call(['centrifuge', 'generatekeys', '-t', 'ed25519', '-p', tpl['keys']['signing']['privateKey'].encode('utf-8'), '-q', tpl['keys']['signing']['publicKey'].encode('utf-8'), '-c', configpath])
        call(['centrifuge', 'generatekeys', '-t', 'ed25519',  tpl['keys']['p2p']['privateKey'].encode('utf-8'), '-q', tpl['keys']['p2p']['publicKey'].encode('utf-8'), '-c', configpath])
        call(['centrifuge', 'addkey', '-p', 'ethauth', '-c', configpath])
        call(['centrifuge', 'addkey', '-p', 'sign', '-c', configpath])
        call(['centrifuge', 'addkey', '-p', 'p2p', '-c', configpath])

def main():
    configpath = createConfigFile()
    id = createIdentity(configpath)
    if id == '':
        print "Error"
        return
    # Update Identity
    with open(configpath, 'r') as stream:
        tpl = yaml.load(stream)
        tpl['identityId'] = id.encode('utf-8')
    with open(configpath, 'w+') as stream:
        yaml.dump(tpl, stream, default_flow_style=False)
    addKeys(configpath)

main()