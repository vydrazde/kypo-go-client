package main

import "kypo-go-client"

func main() {
	client, _ := kypo_go_client.NewClient("https://images.crp.kypo.muni.cz", "bzhwmbxgyxALbAdMjYOgpolQzkiQHGwWRXxm", "kypo-admin", "UfMLMlEw0751kia002Kbv9MaLNlo3T")
	_, _ = client.CreateSandboxDefinition("git@gitlab.ics.muni.cz:muni-kypo-trainings/games/junior-hacker.git", "master")

}
