package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shomali11/slacker"
	"github.com/spf13/viper"
)

var cfgFile = os.Getenv("PINGBOT_CONFIG_FILE")

func main() {
	c := make(chan int)
	go startBot()
	go refreshToken()
	<-c
}

func startBot() {
	bot := slacker.NewClient(viper.GetString("bot-token"), viper.GetString("app-token"))
	definition := &slacker.CommandDefinition{
		Description: "ping command",
		Examples:    []string{"ping"},
		Handler:     pingHandler,
	}
	echoDefinition := &slacker.CommandDefinition{
		Description: "echo command",
		Examples:    []string{"echo <sentence>"},
		Handler:     echoHandler,
	}
	helpDefinition := &slacker.CommandDefinition{
		Description: "Help about any command",
		Handler:     helpLogic(bot),
	}
	bot.Command("ping", definition)
	bot.Command("echo", echoDefinition)
	bot.CustomResponse(NewResponse)
	bot.Help(helpDefinition)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := bot.Listen(ctx); err != nil {
		log.Fatal().Err(err)
	}
}

func echoHandler(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
	log.Printf("echo logs only in debug mode")
	response.Reply(fmt.Sprintf("%s from %s", botCtx.Event().Text, botCtx.Event().UserProfile.Email))
}

func pingHandler(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
	user := botCtx.Event().UserProfile
	reqMsg := botCtx.Event().Text
	log.Info().Msgf("%s from %s", reqMsg, user.Email)
	response.Reply(user.Email)
}

func helpLogic(s *slacker.Slacker) func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
	return func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
		var (
			codeMessageFormat = "```%s```"
			helpMessage       = ""
			dash              = "-"
			space             = " "
			lock              = ":lock:"
			newLine           = "\n"
		)

		for _, command := range s.BotCommands() {
			if command.Definition().HideHelp {
				continue
			}
			tokens := command.Tokenize()
			for _, token := range tokens {
				helpMessage += token.Word + space
			}

			if len(command.Definition().Description) > 0 {
				helpMessage += dash + space + command.Definition().Description
			}

			if command.Definition().AuthorizationFunc != nil {
				helpMessage += space + lock
			}

			helpMessage += newLine

			for _, example := range command.Definition().Examples {
				helpMessage += example + newLine
			}
		}

		response.Reply(fmt.Sprintf(codeMessageFormat, helpMessage))
	}
}

func refreshToken() {
	for {
		log.Printf("refresh token goes here.")
		time.Sleep(time.Second * 5)
	}
}

func init() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			log.Fatal().Err(err)
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(".pingbot")
		viper.SetConfigType("yaml")
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	debug := viper.GetBool("debug")
	log.Info().Msgf("%v", debug)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	}
}
