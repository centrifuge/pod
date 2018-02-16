package keytools

import "io/ioutil"

func writeKeyToFile(fileName string, key []byte) {
	err := ioutil.WriteFile(fileName, key, 0600)
	if err != nil {
		panic(err)
	}
}

