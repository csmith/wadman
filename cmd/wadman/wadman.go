package main

func main() {
	if err := rootCommand.Execute(); err != nil {
		bail("Error executing command: %v", err)
	}
}
