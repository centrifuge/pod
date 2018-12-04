## Welcome to Testworld

`"CHAOS TAKES CONTROL. WELCOME TO THE NEW WORLD. WELCOME TO TESTWORLD"`

Testworld(loosely analogous to [Westworld](https://medium.com/@naveen101/westworld-an-introduction-cc7d29bfbe84) ;) is a simulation and test environment for centrifuge p2p network. 
Here you can create, run and test nodes with various behaviours to observe how they would behave and debug any problems encountered.

### Tutorial 

- All hosts(p2p nodes) in the Testworld are created and maintained during simulation by drFord(`park.go#hostManager`). He also ensures that the hosts and tests are properly cleaned up after each simulation run.
- Bernard(`hostManager.bernard`) is a special host that serves as the libp2p bootnode for the test network.
- `hostConfig` serves as the starting point for you to define new hosts. Please check whether an existing host can be reused for your scenario before adding new ones.
- At the start of the each test run a test config is loaded for the required Ethereum network(eg: Rinkeby or local). The host configs are defined based on this.
- The test initialisation also ensures that geth is running in the background and require smart contracts are migrated for the network.
- Refer `park_test.go` for a simple starting point to define your own simulations/tests.
- Plus points if you write a test with a scenario that matches a scene in Westworld with node names matching the characters ;)




