package prattle

import (
	"github.com/divya2661/prattle/config"
	"github.com/divya2661/prattle/prattle"
)

func main(){
	//TODO: Take file name as user input
	config.Load()
	prattle, err := prattle.NewPrattle(config.SiblingAddr(), config.RpcPort())
}

