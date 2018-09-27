ethereum:
  accounts: 
    main: 
      key: "${ETH_ACCOUNT_KEY}"
      password: "${ETH_ACCOUNT_PWD}"
  nodeURL: "${EHT_NODE_URL}"
identityId: "${IDENTITY}"
keys: 
  ethauth: 
    privateKey: "${DATADIR}/ethauth.key.pem"
    publicKey: "${DATADIR}/ethauth.pub.pem"
  p2p: 
    privateKey: "${DATADIR}/p2p.key.pem"
    publicKey: "${DATADIR}/p2p.pub.pem"
  signing: 
    privateKey: "${DATADIR}/signature.key.pem"
    publicKey: "${DATADIR}/signature.pub.pem"
nodeHostname: "0.0.0.0"
nodePort: "${API_PORT}"
p2p: 
  port: "${P2P_PORT}"
storage: 
  path: "${DATADIR}/db/centrifuge_data.leveldb"
