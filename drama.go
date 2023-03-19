package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

type Words struct {
	Peoples    []string   `json:"peoples"`
	Categories [][]string `json:"categories"`
}

type Config struct {
	Token     string `json:"token"`
	OpenAiKey string `json:"openai_key"`
	GuildID   string `json:"guild_id"`
}

func main() {
	bytes, standardConfig := os.ReadFile("config.json")
	if standardConfig != nil {
		fmt.Println("Error reading file: ", standardConfig)
		return
	}

	var config Config
	standardConfig = json.Unmarshal(bytes, &config)
	if standardConfig != nil {
		fmt.Println("Error unmarshalling JSON: ", standardConfig)
		return
	}

	bytes, wordsConfig := os.ReadFile("words.json")
	if wordsConfig != nil {
		fmt.Println("Error reading file: ", wordsConfig)
		return
	}

	var words Words
	wordsConfig = json.Unmarshal(bytes, &words)
	if wordsConfig != nil {
		fmt.Println("Error unmarshalling JSON: ", wordsConfig)
		return
	}

	discord, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	err = discord.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
		return
	}

	command := &discordgo.ApplicationCommand{
		Name:        "generate",
		Description: "Generuje opis dramatycznej sytuacji",
	}

	_, err = discord.ApplicationCommandCreate(discord.State.User.ID, config.GuildID, command)
	if err != nil {
		fmt.Println("Error creating command: ", err)
		return
	}

	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.ApplicationCommandData().Name == "generate" {
			var wordsArr = [3]string{}

			for i := 0; i < 3; i++ {
				categoryIndex := rand.Intn(len(words.Categories))
				wordIndex := rand.Intn(len(words.Categories[categoryIndex]))
				wordsArr[i] = words.Categories[categoryIndex][wordIndex]
			}

			peopleIndex := rand.Intn(len(words.Peoples))

			prompt := fmt.Sprintf("Proszę wygeneruj dla mnie TYLKO JEDEN opis dramatycznej sytuacji korzystając z wyrazów: %s, pamiętaj aby uwzględnij osobę: %s. Napisz to w jednym krótkim zdaniu.", strings.Join(wordsArr[:], ", "), words.Peoples[peopleIndex])

			client := openai.NewClient(config.OpenAiKey)
			resp, err := client.CreateChatCompletion(
				context.Background(),
				openai.ChatCompletionRequest{
					Model: openai.GPT3Dot5Turbo,
					Messages: []openai.ChatCompletionMessage{
						{
							Role:    openai.ChatMessageRoleUser,
							Content: prompt,
						},
					},
				},
			)

			if err != nil {
				fmt.Printf("Completion error: %v\n", err)
				return
			}

			embed := &discordgo.MessageEmbed{
				Title:       "EternalCode.pl - DramaGPT",
				Description: resp.Choices[0].Message.Content,
				Color:       0x00ff00,
			}

			reply := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
				},
			}

			respond := s.InteractionRespond(i.Interaction, reply)
			if respond != nil {
				fmt.Println(respond)
			}
		}

	})

	fmt.Printf("Bot is now running on %d servers", len(discord.State.Guilds))
	select {}
}
