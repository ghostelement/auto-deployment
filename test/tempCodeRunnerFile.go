func TestYaml(file string) {
	text, err := deploy.Config(file)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(text)
}