let createRoomButton = document.querySelector(".input_create__room")
let joinButtonRoom = document.querySelector(".button__join")
let popup = document.querySelector(".overflow");
let popupButton = document.querySelector(".button__popup");
let tableRoomsButton = document.querySelector(".rooms__container")
let allRoomsButtons = []
let login;

createRoomButton.addEventListener("click", handleCreateRoom)
joinButtonRoom.addEventListener("click",handleJoinRoomByID)
popupButton.addEventListener("click", ()=>{popup.classList.remove("active")})
document.addEventListener("DOMContentLoaded", getAllRooms);


async function getAllRooms() {
    try {
        let response = await fetch("/availablerooms", { method: "GET", credentials: "include" });
        let data = await response.json();
        createAllRoomsButtons(data.login, data.rooms)
        if(allRoomsButtons.length > 0){
            allRoomsButtons.forEach((element)=>{
                element.addEventListener("click", () => handleJoinRoomByButton(element))
            })
        }
    } catch (error) {
        console.error("Ошибка запроса:", error);
    }
}


function createAllRoomsButtons(login, rooms){
    for(let  i = 0; i < rooms.length; i++){
        let button = createButton(login, rooms[i].roomID, rooms[i].playersCnt)
        tableRoomsButton.append(button)
        allRoomsButtons.push(button)
    }
}
function createButton(login, roomId, countPeople){
    const button = document.createElement("a")
    const nameBlock = document.createElement("div")
    const peopleBlock = document.createElement("div")
    const name = document.createElement("span")

    name.classList.add("room__name-different")
    nameBlock.classList.add("room__name")
    peopleBlock.classList.add("room__count")
    button.classList.add("room")

    button.href= '/game?id=' + encodeURIComponent(login) + '&roomId=' + encodeURIComponent(roomId);
    nameBlock.innerText = "name: "
    name.innerText = roomId
    peopleBlock.innerText = `${countPeople}/16`

    nameBlock.append(name)
    button.append(nameBlock)
    button.append(peopleBlock)

    return button;
}

function handleCreateRoom() {
    let roomId = document.querySelector("#createRoom").value
    if (roomId !== null && roomId.trim() !== "") {
        fetch('/createroom', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ roomId: roomId.trim() })
        })
            .then(response => {
                if (!response.ok) {
                    return response.json().then(errData => {
                        throw new Error(errData.error || `Ошибка: ${response.status} ${response.statusText}`);
                    });
                }
                return response.json();
            })
            .then(data => {
                window.location.href = '/game?id=' + encodeURIComponent(data.login) + '&roomId=' + encodeURIComponent(data.roomId);
            })
            .catch(error => {
                popup.classList.add("active")
                popupButton.innerText = "Продолжить"
                popup.querySelector(".popup__text").innerText = error.message
            });
    } else{
        popup.classList.add("active")
        popupButton.innerText = "Продолжить"
        popup.querySelector(".popup__text").innerText = "Некорректный id комнаты"
        return;
    }
}
function handleJoinRoomByButton(element){
    let roomId = element.querySelector(".room__name-different").innerText
    if(roomId != null){
        const url = `/joinroombyid`;
        const bodyData = {
            roomId: roomId.trim()
        };
        fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(bodyData)
        })
            .then(response => {
                if (!response.ok) {
                    return response.json().then(errData => {
                        throw new Error(errData.error || `Ошибка: ${response.status} ${response.statusText}`);
                    });
                }
                return response.json();
            })
            .then(data => {
                window.location.href = '/game?id=' + encodeURIComponent(data.login) + '&roomId=' + encodeURIComponent(data.roomId);
            })
            .catch(error => {
                popup.classList.add("active")
                popupButton.innerText = "Продолжить"
                popup.querySelector(".popup__text").innerText = error.message
            });
    }
    
}
function handleJoinRoomByID() {
    const roomId = document.querySelector("#joinRoom").value
    if (!roomId || roomId.trim() === "") {
        popup.classList.add("active")
        popupButton.innerText = "Продолжить"
        popup.querySelector(".popup__text").innerText = "Некорректный id комнаты"
        return;
    }
    const url = `/joinroombyid`;
    const bodyData = {
        roomId: roomId.trim()
    };
    fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(bodyData)
    })
        .then(response => {
            if (!response.ok) {
                return response.json().then(errData => {
                    throw new Error(errData.error || `Ошибка: ${response.status} ${response.statusText}`);
                });
            }
            return response.json();
        })
        .then(data => {
            window.location.href = '/game?id=' + encodeURIComponent(data.login) + '&roomId=' + encodeURIComponent(data.roomId);
        })
        .catch(error => {
            popup.classList.add("active")
            popupButton.innerText = "Продолжить"
            popup.querySelector(".popup__text").innerText = error.message
        });
}