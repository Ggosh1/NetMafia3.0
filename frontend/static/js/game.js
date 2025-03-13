let noBtn = document.querySelector(".button__popup-no")
let yesBtn = document.querySelector(".button__popup-yes")
let popup = document.querySelector(".overflow")


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
            alert("Ошибка при выходе: " + error.message);
        });
}