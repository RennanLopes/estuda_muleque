package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows/registry"
	"google.golang.org/api/option"
)

// Estrutura App
type App struct {
	ctx      context.Context
	finished bool
}

// NewApp cria uma nova instância da estrutura App

func NewApp() *App {
	return &App{
		finished: false,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	rand.Seed(time.Now().UnixNano())
	a.RegisterStartup()
}

func (a *App) RegisterStartup() {
	execPath, err := os.Executable()
	if err != nil {
		return
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return
	}
	defer key.Close()

	key.SetStringValue("EstudaMuleque", execPath)
}

// beforeClose é chamado quando o usuário tenta fechar o app
func (a *App) beforeClose(ctx context.Context) bool {
	if !a.finished {
		runtime.EventsEmit(ctx, "show_warning", "Você precisa terminar o desafio primeiro!")
		return true // Bloqueia o fechamento
	}
	return false // Permite o fechamento
}

// FinishChallenge é chamado pelo frontend quando as perguntas são respondidas
func (a *App) FinishChallenge() {
	a.finished = true
	runtime.WindowSetAlwaysOnTop(a.ctx, false)
	runtime.Quit(a.ctx)
}

type Question struct {
	Text    string   `json:"question"`
	Options []string `json:"options"`
	Answer  int      `json:"answer"`
	Subject string   `json:"subject"`
}

// GetQuestions busca perguntas do Google Gemini ou retorna as padrões
func (a *App) GetQuestions(count int) ([]Question, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")

	// Tenta carregar do config.json se a variável de ambiente estiver vazia
	if apiKey == "" {
		// Procura pelo config.json no mesmo diretório do executável
		exePath, _ := os.Executable()
		configPath := filepath.Join(filepath.Dir(exePath), "config.json")
		fmt.Printf("Buscando config em: %s\n", configPath)

		if data, err := os.ReadFile(configPath); err == nil {
			var config struct {
				GeminiKey string `json:"gemini_key"`
				APIKey    string `json:"api_key"`
				OpenAIKey string `json:"openai_key"`
			}
			if err := json.Unmarshal(data, &config); err == nil {
				if config.GeminiKey != "" {
					apiKey = config.GeminiKey
					fmt.Println("Chave Gemini carregada do config da aplicação.")
				} else if config.APIKey != "" {
					apiKey = config.APIKey
					fmt.Println("Chave API carregada do config da aplicação.")
				} else if config.OpenAIKey != "" {
					apiKey = config.OpenAIKey
					fmt.Println("Chave OpenAI carregada do config da aplicação.")
				}
			}
		} else {
			// Fallback para o diretório atual em desenvolvimento
			fmt.Println("Config não encontrado no diretório do executável, tentando diretório atual...")
			if data, err := os.ReadFile("config.json"); err == nil {
				var config struct {
					GeminiKey string `json:"gemini_key"`
					APIKey    string `json:"api_key"`
					OpenAIKey string `json:"openai_key"`
				}
				if err := json.Unmarshal(data, &config); err == nil {
					if config.GeminiKey != "" {
						apiKey = config.GeminiKey
						fmt.Println("Chave Gemini carregada do diretório atual.")
					} else if config.APIKey != "" {
						apiKey = config.APIKey
						fmt.Println("Chave API carregada do diretório atual.")
					} else if config.OpenAIKey != "" {
						apiKey = config.OpenAIKey
						fmt.Println("Chave OpenAI carregada do diretório atual.")
					}
				}
			}
		}
	}

	if apiKey == "" {
		fmt.Println("CRITICAL ERROR: No API Key found after checking all sources!")
		return getDummyQuestions(count), nil
	}

	fmt.Printf("API Key identified (first 5 chars): %s...\n", apiKey[:5])

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		fmt.Printf("ERROR: Failed to create Gemini client: %v\n", err)
		return getDummyQuestions(count), nil
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.5-flash")
	model.SetTemperature(0.7)
	model.ResponseMIMEType = "application/json"

	prompt := fmt.Sprintf(`Gere %d perguntas de múltipla escolha para uma criança de 8 anos em Português. 
	As matérias devem ser: Matemática, Ciências, História, Inglês e Curiosidades.
	O nível de dificuldade deve ser fácil a médio.
	Retorne um objeto JSON com uma chave "questions" contendo um array de objetos com os campos: "question" (string), "options" (array de 4 strings), "answer" (index 0-3 da opção correta) e "subject" (string).`, count)

	fmt.Println("Calling Gemini API...")
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		fmt.Printf("ERROR: Gemini API call failed: %v\n", err)
		return getDummyQuestions(count), nil
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		fmt.Println("ERROR: Empty response candidates from Gemini")
		return getDummyQuestions(count), nil
	}

	var sb strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			sb.WriteString(string(text))
		}
	}

	jsonStr := sb.String()
	fmt.Printf("DEBUG: Received JSON string from Gemini (length: %d)\n", len(jsonStr))

	// Limpa blocos de código markdown se presentes
	jsonStr = strings.TrimSpace(jsonStr)
	if strings.HasPrefix(jsonStr, "```json") {
		jsonStr = strings.TrimPrefix(jsonStr, "```json")
		jsonStr = strings.TrimSuffix(jsonStr, "```")
	} else if strings.HasPrefix(jsonStr, "```") {
		jsonStr = strings.TrimPrefix(jsonStr, "```")
		jsonStr = strings.TrimSuffix(jsonStr, "```")
	}
	jsonStr = strings.TrimSpace(jsonStr)

	var result struct {
		Questions []Question `json:"questions"`
	}
	err = json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		fmt.Printf("ERROR: JSON Unmarshal failed: %v\n", err)
		fmt.Printf("Raw response part: %s\n", jsonStr)
		return getDummyQuestions(count), nil
	}

	if len(result.Questions) == 0 {
		fmt.Println("ERROR: Resulting questions array is empty")
		return getDummyQuestions(count), nil
	}

	fmt.Printf("SUCCESS: Loaded %d questions from Gemini\n", len(result.Questions))
	return result.Questions, nil
}

func getDummyQuestions(count int) []Question {
	allQuestions := []Question{
		{Text: "Quanto é 10 + 5?", Options: []string{"12", "15", "18", "20"}, Answer: 1, Subject: "Matemática"},
		{Text: "Qual animal é conhecido como o rei da selva?", Options: []string{"Tigre", "Elefante", "Leão", "Zebra"}, Answer: 2, Subject: "Ciências"},
		{Text: "Como se diz 'Obrigado' em inglês?", Options: []string{"Hello", "Please", "Thank you", "Goodbye"}, Answer: 2, Subject: "Inglês"},
		{Text: "Quem descobriu o Brasil?", Options: []string{"Dom Pedro I", "Pedro Álvares Cabral", "Cristóvão Colombo", "Dilma Rousseff"}, Answer: 1, Subject: "História"},
		{Text: "Qual planeta é conhecido como o planeta vermelho?", Options: []string{"Vênus", "Marte", "Júpiter", "Saturno"}, Answer: 1, Subject: "Ciências"},
		{Text: "Quantos dias tem uma semana?", Options: []string{"5", "6", "7", "8"}, Answer: 2, Subject: "Geral"},
		{Text: "Qual é o maior mamífero do mundo?", Options: []string{"Elefante", "Baleia Azul", "Girafa", "Tubarão"}, Answer: 1, Subject: "Ciências"},
		{Text: "Qual é o nome do satélite natural da Terra?", Options: []string{"Sol", "Marte", "Lua", "Estrela"}, Answer: 2, Subject: "Ciências"},
		{Text: "Quanto é 2 x 5?", Options: []string{"7", "8", "10", "12"}, Answer: 2, Subject: "Matemática"},
		{Text: "Qual é a primeira letra do alfabeto?", Options: []string{"B", "C", "A", "Z"}, Answer: 2, Subject: "Português"},
		{Text: "Qual fruta é amarela e macaco adora?", Options: []string{"Maçã", "Uva", "Banana", "Pera"}, Answer: 2, Subject: "Geral"},
		{Text: "Como se chama o bebê da galinha?", Options: []string{"Leitão", "Pinto", "Bezerro", "Cordeiro"}, Answer: 1, Subject: "Ciências"},
		{Text: "Qual é a cor da grama?", Options: []string{"Azul", "Amarelo", "Verde", "Roxo"}, Answer: 2, Subject: "Ciências"},
		{Text: "O que usamos para ver as horas?", Options: []string{"Sapato", "Relógio", "Colher", "Meia"}, Answer: 1, Subject: "Geral"},
		{Text: "Qual é a maior estrela do nosso sistema?", Options: []string{"Terra", "Lua", "Sol", "Marte"}, Answer: 2, Subject: "Ciências"},
		{Text: "Qual o nome do país onde moramos?", Options: []string{"Brasil", "EUA", "França", "Japão"}, Answer: 0, Subject: "Geral"},
		{Text: "Quantas patas tem uma aranha?", Options: []string{"4", "6", "8", "10"}, Answer: 2, Subject: "Ciências"},
		{Text: "O que o gelo vira quando derrete?", Options: []string{"Fogo", "Pedra", "Água", "Vapor"}, Answer: 2, Subject: "Ciências"},
		{Text: "Quanto é 20 - 7?", Options: []string{"11", "12", "13", "14"}, Answer: 2, Subject: "Matemática"},
		{Text: "Qual é a cor do céu em um dia ensolarado?", Options: []string{"Verde", "Azul", "Vermelho", "Preto"}, Answer: 1, Subject: "Ciências"},
		{Text: "Qual é o oposto de 'alto'?", Options: []string{"Grande", "Largo", "Baixo", "Fino"}, Answer: 2, Subject: "Português"},
		{Text: "Qual animal faz 'Muuu'?", Options: []string{"Gato", "Cachorro", "Vaca", "Cavalo"}, Answer: 2, Subject: "Geral"},
		{Text: "Qual é a capital do Brasil?", Options: []string{"Rio de Janeiro", "São Paulo", "Brasília", "Salvador"}, Answer: 2, Subject: "História"},
		{Text: "Quantos meses tem um ano?", Options: []string{"10", "11", "12", "13"}, Answer: 2, Subject: "Geral"},
		{Text: "Qual é a cor da banana?", Options: []string{"Vermelha", "Azul", "Amarela", "Verde"}, Answer: 2, Subject: "Geral"},
		{Text: "Quem pintou a Mona Lisa?", Options: []string{"Picasso", "Van Gogh", "Leonardo da Vinci", "Monet"}, Answer: 2, Subject: "História"},
		{Text: "Qual é o maior continente do mundo?", Options: []string{"África", "Ásia", "Europa", "América"}, Answer: 1, Subject: "Geografia"},
		{Text: "Quantos oceanos existem na Terra?", Options: []string{"3", "4", "5", "6"}, Answer: 2, Subject: "Geografia"},
		{Text: "Qual é o metal mais precioso?", Options: []string{"Prata", "Ouro", "Bronze", "Ferro"}, Answer: 1, Subject: "Curiosidades"},
		{Text: "Qual é o idioma oficial da China?", Options: []string{"Inglês", "Chinês", "Espanhol", "Francês"}, Answer: 1, Subject: "Geral"},
		{Text: "Qual é a montanha mais alta do mundo?", Options: []string{"K2", "Everest", "Fuji", "Andes"}, Answer: 1, Subject: "Geografia"},
		{Text: "Qual é a cor da neve?", Options: []string{"Branca", "Azul", "Cinza", "Transparente"}, Answer: 0, Subject: "Ciências"},
		{Text: "Qual é o animal mais rápido do mundo?", Options: []string{"Leão", "Guepardo", "Cavalo", "Águia"}, Answer: 1, Subject: "Ciências"},
		{Text: "Quem inventou a lâmpada?", Options: []string{"Isaac Newton", "Albert Einstein", "Thomas Edison", "Graham Bell"}, Answer: 2, Subject: "História"},
		{Text: "Qual é a capital da França?", Options: []string{"Londres", "Berlim", "Paris", "Roma"}, Answer: 2, Subject: "Geografia"},
	}

	// Embaralha
	rand.Shuffle(len(allQuestions), func(i, j int) {
		allQuestions[i], allQuestions[j] = allQuestions[j], allQuestions[i]
	})

	if count > len(allQuestions) {
		count = len(allQuestions)
	}

	return allQuestions[:count]
}
