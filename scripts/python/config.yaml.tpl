storage:
  path: ${DATADIR}/db/centrifuge_data.leveldb
nodeHostname: 0.0.0.0
nodePort: ${API_PORT}
p2p:
  port: ${P2P_PORT}
identityId: ${IDENTITY}
keys:
  p2p:
    publicKey: ${DATADIR}/p2p.pub.pem
    privateKey: ${DATADIR}/p2p.key.pem
  signing:
    publicKey: ${DATADIR}/signature.pub.pem
    privateKey: ${DATADIR}/signature.key.pem
  ethauth:
    publicKey: ${DATADIR}/ethauth.pub.pem
    privateKey: ${DATADIR}/ethauth.key.pem

ethereum:
  nodeURL: ${EHT_NODE_URL}
  accounts:
    main:
      key: '${ETH_ACCOUNT_KEY}'
      password: '${ETH_ACCOUNT_PWD}'
