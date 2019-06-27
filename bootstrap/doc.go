/*

Package bootstrap maintains the bootstrapping interface design for go-centrifuge.
These interfaces are defined to enable decoupling between implementation of various implementation packages of
go-centrifuge and their runtime instantiation logic.

Why Bootstrapper interface was needed

Unlike a regular library, process runtime code has certain dependencies on a central instantiation
point in the software for it to be initialised properly. For go-centrifuge this made it difficult to
change instantiation logic based on the need(eg: commands vs daemon) because of cyclic dependencies or to reuse that logic for purposes of testing.
The use of global variables to access already initialised implementations meant that it was hard to identify the dependencies needed for each module to function properly.
Also this meant that instantiation logic got coupled to a particular implementation of a package interface. For example leveldb implementation of storage
was being used even for unit tests because there was no way to mock it out individually.

Other than solving those problems while allowing the code to slowly evolve, Bootstrappers satisfies the need for inversion of control(IOC)
where modules of code get injected with their dependencies based on a requirement defined by each package in the bootstrapper.go.
This convention of defining a bootstrapper.go in each package also serves as a good documentation source as well as to reduce
complexity in understanding the code base at a global level. The package level bootstrapper also helps in implementing runtime
behavior modifications based on IOC, that are very useful in simulations.

How Bootstrappers work

The interface is simple it has a simple function,

	Bootstrap(context map[string]interface{}) error

This can be implemented by each package that needs runtime instantiation based on some dependencies that are already defined in the runtime.
The bootstrapper calling sequence is hard coded in for the particular scenario(with mocks if needed in a test) and called by the central instantiation point(main).
The context map contains objects already defined in the runtime keyed by well known names defined with constants in each package that exposes those objects interfaces.

Eg:
First storage package defines an interface and a key for the implementation in the context map in storage/repo.go

	type Repository {
		Get(key []byte) []byte
	}

	const BootstrappedDB string = "BootstrappedDB"

Then the leveldb implementation of the above interface in leveldb package injects the instantiated leveldb implementation in storage/leveldb/bootstrapper.go

	// ..
	func Bootstrap(context map[string]interface{}) error {
		// ...

		context[storage.BootstrappedDB] = NewLevelDB(..)
	}

After this a downstream package bootstrapped can use the context[storage.BootstrappedDB] object to for its own instantiation, in document/bootstrapper.go

	// ..
	ldb, ok := ctx[storage.BootstrappedDB].(storage.Repository)
	if !ok {
		return ErrDocumentBootstrap
	}

	repo := NewDocRepository(ldb)

Check go-centrifuge/cmd package to see how the Bootstrapper interfaces are used to bootstrap go-centrifuge.
*/
package bootstrap
