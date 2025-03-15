let noBtn = document.querySelector(".button__popup-no")
let yesBtn = document.querySelector(".button__popup-yes")
let popup = document.querySelector(".overflow")
let popup1 = document.querySelector(".overflow1")
let ws;
let timerInterval;
let selectedPlayerID = null; // Хранит ID выбранного игрока
let needToChooseTarget = false; // Флаг, показывающий необходимость выбора цели
let chooseTargetMode = false; // Флаг, показывающий режим выбора
let targetPlayerID = "";
let flad = true
const players = {}; // Информация об игроках
let currentVotes = {};
let playersVotedFor = {};
const params = new URLSearchParams(window.location.search);
const playerId = params.get("id");
let endGame = false;
let mafia = {};
if (!playerId) {
    alert("ID игрока не указан. Пожалуйста, заходите через страницу входа.");
}

noBtn.addEventListener("click",()=>{
    popup.classList.remove("active")
})
yesBtn.addEventListener("click",()=>{
    leaveRoom();
})



function handleExit() {
    const params = new URLSearchParams(window.location.search);
    const username = params.get("id");
    if (!username) {
        alert("Ошибка: имя пользователя не указано.");
        return;
    }
    popup.classList.add("active")
}


function leaveRoom(){
    //log("Пытаюсь выйти из игры")
    const params = new URLSearchParams(window.location.search);
    const username = params.get("id");
    const roomId = params.get("roomId");
    fetch(`/leaveroom?id=${encodeURIComponent(username)}&roomId=${encodeURIComponent(roomId)}`)
        .then(response => {
            //log("Хуйня какая-то")
            if (!response.ok) {
                return response.text().then(text => { throw new Error(text); });
            }
            return response.json();
        })
        .then(data => {
            log("Выход выполнен успешно:", data);
            log("Выход выполнен успешно:");
            window.location.href = '/profile?id=' + encodeURIComponent(username);
        })
        .catch(error => {
            log("Ошибка при выходе из комнаты:", error);
            popup1.classList.add("active")
            popup1.querySelector(".popup__text").innerHTML = error.message
            popup1.querySelector(".button__popup").innerHTML = "Продолжить"
        });
}


const sun = document.querySelector('.sun');
const moon = document.querySelector('.moon');
let sunAngle = 90; 
let moonAngle = 270; 

function moveSunMoon(clockwise) {
    sun.classList.remove('moveClockwise', 'moveCounterClockwise');
    moon.classList.remove('moveClockwise', 'moveCounterClockwise');
    sunAngle += 180; 
    moonAngle += 180; 
    if(sun.classList.contains("onTop")){
        sun.style.transform = `translate(0%, -90%) rotate(${sunAngle}deg) translateX(-1700px) rotate(${sunAngle}deg)`;
        moon.style.transform = `translate(0%, -90%) rotate(${moonAngle}deg) translateX(-250px) rotate(${moonAngle}deg)`;
        sun.classList.remove("onTop")
    }else{
        sun.style.transform = `translate(0%, -90%) rotate(${sunAngle}deg) translateX(-250px) rotate(${sunAngle}deg)`;
        moon.style.transform = `translate(0%, -90%) rotate(${moonAngle}deg) translateX(-1700px) rotate(${moonAngle}deg)`;
        sun.classList.add("onTop")
    }
}


// Обработчик нажатия Enter для чата
const input = document.getElementById("chat-input");
input.addEventListener("keydown", function(event) {
    if (event.key === "Enter") {
        sendChatMessage();
    }
});


// Закрытие WebSocket при закрытии вкладки
window.addEventListener("beforeunload", function () {
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.close(500, "Tab closed");
    }
});


connect();

function connect() {
    if (!playerId) {
        log("ID игрока не найден. Зайдите через страницу входа.");
        return;
    }
    ws = new WebSocket(`ws://158.160.138.22:80/ws?id=${encodeURIComponent(playerId)}`);
    ws.onopen = () => { log(`Connected as ${playerId}`); };
    ws.onmessage = (event) => {
        const message = JSON.parse(event.data);
        handleServerMessage(message);
    };
    ws.onclose = () => { log("Disconnected from server"); };
}

let isGameover = false

function handleServerMessage(message) {
    if (isGameover) {
        return
    }

    if (message.time_remaining !== undefined) {
        document.getElementById("time-remaining").textContent = message.time_remaining;
    }

    selectedPlayerID = message.player_vote;
    targetPlayerID = message.target;

    // Если сервер сообщает, что нужно выбрать цель, обновляем флаг и отображение кнопок
    if (message.need_to_choose_target !== undefined) {
        needToChooseTarget = message.need_to_choose_target;
        const targetControls = document.getElementById("target-controls");
        targetControls.style.display = needToChooseTarget ? "block" : "none";
    }

    if (message.role) {
        updateRoleDisplay(message.role);
    }

    if (message.mafia_list){
        //log(message.mafia_list)
        mafia = message.mafia_list
    }

    if (message.players_voted_for) {
        playersVotedFor = message.players_voted_for;
        updatePlayerGrid(players);
    }

    if (message.error) {
        popup1.classList.add("active")
        popup1.querySelector(".popup__text").innerText = message.error
        popup1.querySelector(".button__popup").innerText = "Продолжить"
        popup1.querySelector(".button__popup").addEventListener("click",()=>{
            popup1.classList.remove("active")
        })
    }
    
    if (message.phase) {
        if (message.phase === "night") {
            if(flad){
                moveSunMoon(true)
                flad = false
                document.querySelector("body").classList.add("active")
            } 
        } else if (message.phase === "day") {
            if(!flad){
                flad = true
                moveSunMoon(true)
                document.querySelector("body").classList.remove("active")
            }
        } else if (message.phase === "gameover") {
            document.querySelector("body").classList.remove("active")
            isGameover = true
            popup1.classList.add("active")
            popup1.querySelector(".popup__text").innerText = message.winner
            popup1.querySelector(".button__popup").innerText = "Продолжить"
            popup1.querySelector(".button__popup").addEventListener("click",leaveRoom)
        }
    }

    if (message.players) {
        updatePlayerGrid(message.players);
    }

    /*if (message.winner) {
        log(`Game over! ${message.winner}`);
        alert(`Game over! ${message.winner}`);
        leaveRoom()
    }*/

    if (message.chat) {
        displayChatMessage(message.playerID, message.chat);
    }

    // Обработка истории чата
    if (message.type && message.type === "chatHistory" && message.history) {
        message.history.forEach(msg => {
            displayChatMessage(msg.playerID, msg.chat);
        });
    }

    if (message.votes) {
        currentVotes = message.votes;
        updatePlayerGrid(players);
    }

    if (message.type && message.type === "playerList") {
        updatePlayerGrid(message.players);
        return;
    }
}

function updatePlayerGrid(playerStatus) {
    const playerGrid = document.getElementById("player-grid");
    playerGrid.innerHTML = "";
    Object.entries(playerStatus).forEach(([playerID, isAlive]) => {
        const playerCard = document.createElement("div");
        playerCard.classList.add("player-card");
        playerCard.classList.add(isAlive ? "alive" : "dead");

        if (selectedPlayerID === playerID && isAlive) {
            playerCard.classList.add("selected");
        }
        playerCard.innerHTML = `
        <div class="player-avatar"></div>
        <div class="player-name">${playerID}</div>
        <div>Status: ${isAlive ? "Alive" : "Dead"}</div>
    `;
        const votedFor = currentVotes[playerID];
        if (votedFor) {
            const voteSign = document.createElement("div");
            voteSign.classList.add("vote-sign");
            voteSign.textContent = "⚔️ " + votedFor + " ⚔️";
            playerCard.appendChild(voteSign);
        }
        if (targetPlayerID === playerID){
            const targetSign = document.createElement("div");
            targetSign.classList.add("target-sign");
            targetSign.textContent = "🎯";
            playerCard.appendChild(targetSign);
        }
        if (isAlive) {
            playerCard.onclick = () => handleVoteClick(playerID, isAlive);
        }
        if (mafia.includes(playerID)){
            const wolfSign = document.createElement("div");
            wolfSign.classList.add("isMafia");
            wolfSign.textContent = "🐺";
            playerCard.appendChild(wolfSign);
        }
        if (playerID in playersVotedFor){
            const voteSign = document.createElement("div");
            voteSign.classList.add("VotedFor");
            voteSign.textContent = "voted for: " + playersVotedFor[playerID].toString()
            playerCard.appendChild(voteSign);
        }
        playerGrid.appendChild(playerCard);
        players[playerID] = isAlive;
    });
}

function log(message) {
    const logDiv = document.getElementById("log");
    const p = document.createElement("p");
    p.textContent = message;
    logDiv.appendChild(p);
    logDiv.scrollTop = logDiv.scrollHeight;
}

function startGame() {
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ action: "start_game" }));
        log("Game start request sent to the server.");
    } else {
        log("WebSocket is not connected");
    }
}

// Функция для выбора цели
function ChooseTarget() {
    if (ws && ws.readyState === WebSocket.OPEN) {
        chooseTargetMode = true;
        log("Режим выбора цели активирован. Нажмите на игрока для выбора цели.");
        // Если есть контейнер с кнопками выбора, можно его показать:
        //document.getElementById("target-controls").style.display = "block";
    } else {
        log("WebSocket не подключён.");
    }
}

function updateRoleDisplay(role) {
    const roleDisplay = document.getElementById("role-display");
    roleDisplay.textContent = `Your role: ${role}`;
}

function sendChatMessage() {
    const input = document.getElementById("chat-input");
    const message = input.value.trim();
    if (message && ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ action: "chat", message }));
        input.value = "";
    } else if (!ws || ws.readyState !== WebSocket.OPEN) {
        log("WebSocket is not connected.");
    }
}

function displayChatMessage(playerID, chatMessage) {
    const logDiv = document.getElementById("log");
    const p = document.createElement("p");
    p.innerHTML = `<strong>${playerID}:</strong> ${chatMessage}`;
    logDiv.appendChild(p);
    logDiv.scrollTop = logDiv.scrollHeight;
}

function handleVoteClick(playerID, isAlive) {
    if (ws && ws.readyState === WebSocket.OPEN) {
        if (chooseTargetMode) {
            // Если активирован режим выбора цели – отправляем выбор цели
            ws.send(JSON.stringify({ action: "choose_target", target: playerID }));
            log(`Выбрана цель: ${playerID}`);
            chooseTargetMode = false; // Выходим из режима выбора цели
            //document.getElementById("target-controls").style.display = "none";
        } else {
            // Обычная логика голосования
            ws.send(JSON.stringify({ action: "vote", vote: playerID }));
        }
        updatePlayerGrid(players);
    } else {
        log("WebSocket не подключён.");
    }
}
