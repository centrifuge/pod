## Welcome to Testworld

`"CHAOS TAKES CONTROL. WELCOME TO THE NEW WORLD. WELCOME TO TESTWORLD"`

Testworld (loosely analogous to [Westworld](https://medium.com/@naveen101/westworld-an-introduction-cc7d29bfbe84) ;) is a simulation and test environment for centrifuge p2p network. 
Here you can create, run and test nodes with various behaviours to observe how they would behave and debug any problems encountered.

### Tutorial 

- All hosts (p2p nodes) in the Testworld are created and maintained during simulation by drFord(`park.go#hostManager`). He also ensures that the hosts and tests are properly cleaned up after each simulation run.
- Bernard (`hostManager.bernard`) is a special host that serves as the libp2p bootnode for the test network.
- `hostConfig` serves as the starting point for you to define new hosts. Please check whether an existing host can be reused for your scenario before adding new ones.
- At the start of the each test run a test config is loaded for the required Ethereum network(eg: Rinkeby or local). The host configs are defined based on this.
- If you want to define a custom configuration(s) for your local testing copy `configs/base.json` to `configs/local/local.json` and modify the file the way you want. Testworld will automatically pick up the file. 
- You could also replace local config depending you requirements without having to change code by changing `otherLocalConfig` field in `configs/local/local.json`, this is useful when you want to have multiple configs locally for different centrifuge/ethereum networks such as Kovan or Rinkeby. Example:
    in `configs/local/local.json` add the following (please remove comments),
    ```
    {
      "otherLocalConfig": "configs/local/kovan.json",
      
      // following doesn't matter as those would be ignore because of first line
      "runPOAGeth": true,
      "createHostConfigs": true,
      "runMigrations": false,
      "ethNodeURL": "",
      "accountKeyPath": "../build/scripts/test-dependencies/test-ethereum/migrateAccount.json",
      "accountPassword": "",
      "network" : "testing",
      "txPoolAccess": true
    }
    ```
    in `configs/local/kovan.json` add the following (please remove comments),
    ```
    {
      "otherLocalConfig": "",
      "runPOAGeth": true,
      "createHostConfigs": true,
      "runMigrations": false,
      "ethNodeURL": "ws://127.0.0.1:9547",
      "accountKeyPath": "<kovan account>",
      "accountPassword": "",
      
      // bernalheights is the Centrifuge network on Kovan
      "network" : "bernalheights",
      "txPoolAccess": true
    }
    ```
- The test initialisation also ensures that geth is running in the background and required smart contracts are migrated for the network.
- Refer `park_test.go` for a simple starting point to define your own simulations/tests.
- Each test scenario must be defined in a `testworld/<scenario_name>_test.go` file and the build tag `// +build testworld` must be included at the top.
- Plus points if you write a test with a scenario that matches a scene in Westworld with node names matching the characters ;)

### Dev
#### Speed improvements for local testing
At `configs/local/local.json`,
- Set `runMigrations` to `false` after the contracts are deployed once at the local geth node.
- Set `createHostConfigs` to `false` after configs have been generated in `hostconfigs` dir, note that if you add new hosts using `hostConfig` you would need this to be set to `true` again to generate the config for the new host.

