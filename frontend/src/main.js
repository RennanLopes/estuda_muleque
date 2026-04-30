import './style.css';
import { GetQuestions, FinishChallenge } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

let questions = [];
let currentIndex = 0;
let totalScore = 0;
const TARGET_SCORE = 15;

const app = document.getElementById('app');

function renderWarning(msg) {
    const warning = document.createElement('div');
    warning.className = 'warning';
    warning.innerText = msg;
    document.body.appendChild(warning);
    setTimeout(() => warning.remove(), 3000);
}

EventsOn('show_warning', (msg) => {
    renderWarning(msg);
});

async function init() {
    await startRound(20);
}

async function startRound(count) {
    app.innerHTML = '<div class="card"><h1>Carregando seu desafio...</h1><p>Prepare-se!</p></div>';
    try {
        questions = await GetQuestions(count);
        if (!questions || questions.length === 0) {
            throw new Error("Nenhuma pergunta encontrada");
        }
        currentIndex = 0;
        showQuestion();
    } catch (e) {
        console.error(e);
        app.innerHTML = '<div class="card"><h1>Erro ao carregar perguntas.</h1><p>Verifique sua conexão.</p></div>';
    }
}

function showQuestion() {
    if (currentIndex >= questions.length) {
        showResults();
        return;
    }

    const q = questions[currentIndex];
    app.innerHTML = `
        <div class="card">
            <div class="subject">${q.subject}</div>
            <div class="question-text">${q.question}</div>
            <div class="options">
                ${q.options.map((opt, i) => `<button id="opt-${i}">${opt}</button>`).join('')}
            </div>
            <div class="progress">Pergunta ${currentIndex + 1} de ${questions.length} | Acertos totais: ${totalScore}</div>
        </div>
    `;

    q.options.forEach((_, i) => {
        document.getElementById(`opt-${i}`).onclick = () => checkAnswer(i);
    });
}

function checkAnswer(index) {
    const q = questions[currentIndex];
    const buttons = document.querySelectorAll('.options button');
    
    if (index === q.answer) {
        buttons[index].classList.add('correct');
        totalScore++;
    } else {
        buttons[index].classList.add('wrong');
        buttons[q.answer].classList.add('correct');
    }

    buttons.forEach(b => b.disabled = true);

    setTimeout(() => {
        currentIndex++;
        showQuestion();
    }, 1500);
}

function showResults() {
    if (totalScore >= TARGET_SCORE) {
        app.innerHTML = `
            <div class="card">
                <h1>Parabéns!</h1>
                <p>Você completou o desafio de hoje!</p>
                <p>Você acertou no total ${totalScore} perguntas!</p>
                <button id="finish-btn">Liberar Computador</button>
            </div>
        `;
        document.getElementById('finish-btn').onclick = () => {
            FinishChallenge();
        };
    } else {
        app.innerHTML = `
            <div class="card">
                <h1>Quase lá!</h1>
                <p>Você acertou ${totalScore} perguntas, mas precisa de ${TARGET_SCORE} para liberar o computador.</p>
                <p>Vamos tentar mais 15 perguntas?</p>
                <button id="more-btn">Continuar Desafio</button>
            </div>
        `;
        document.getElementById('more-btn').onclick = () => {
            startRound(15);
        };
    }
}

init();
