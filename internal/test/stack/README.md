# Stack Testing
A stack is running Fractal Engine + all dependancies.

- This is achieved by configuring dockerfiles and a docker compose.
- We also have shell scripts so that we can run multiple stacks at the same time without conflicts so that they can be peered.

## Running stacks
./scripts/run_stack.sh 1
./scripts/run_stack.sh 2

Will spin up two entire stacks

## Running stack tests
go test -v ./internal/test/stack/stack_test.go

- This will connect to the stacks + retrieve configuration for ports/keys.
- This will also peer the DogeNet peers and the Dogecoin RPCs.
- It will then run through a series of real world tests.

## Note
The advantage of this setup is it will be as close as possible to a real world test because all the apps are built and deployed.
