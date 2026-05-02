# Estuda Muleque! 🚀

Uma ferramenta educacional que incentiva o estudo antes da diversão.
Seu filho gosta de jogar no PC? Com o **Estuda Muleque**, ele só joga depois de provar que estudou!

O **Estuda Muleque** foi criado para transformar o tempo de tela em uma oportunidade de aprendizado. Utilizando Inteligência Artificial, o sistema gera desafios personalizados que precisam ser resolvidos para que o acesso ao sistema operacional seja liberado.

## Como funciona?
- O programa abre automaticamente quando o usuário faz login no Windows.
- Ele entra em modo tela cheia e fica "sempre no topo", impedindo o acesso a outros programas.
- O seu filho deve responder 30 perguntas de múltipla escolha.
- As matérias incluem: Matemática, Ciências, História, Inglês e Curiosidades.
- O computador só é liberado após a conclusão do desafio.

## Configuração para Desenvolvedores
Para rodar o projeto localmente ou gerar o executável:

1.  **Pré-requisitos:**
    *   [Go](https://golang.org/dl/) (1.18+)
    *   [Node.js](https://nodejs.org/) e NPM
    *   [Wails CLI](https://wails.io/docs/gettingstarted/installation) (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

2.  **Configuração da API:**
    *   Obtenha uma chave no [Google AI Studio](https://aistudio.google.com/).
    *   Crie o arquivo `build/bin/config.json` cole o texto abaixo com sua chave:
    ```json
    {
      "gemini_key": "SUA_CHAVE_AQUI"
    }
    ```

3.  **Execução em Desenvolvimento:**
    ```bash
    wails dev
    ```

4.  **Gerar o Build Final:**
    ```bash
    wails build
    ```
    O executável será gerado em `build/bin/`.

## Tecnologias utilizadas
- **Go (Golang)**: Backend e lógica de bloqueio.
- **Wails**: Interface grafica moderna e simples!
- **Google Gemini API (2.5 Flash)**: Geração dinâmica de perguntas com inteligência artificial de última geração.
- **Windows Registry**: Persistência de inicialização.
