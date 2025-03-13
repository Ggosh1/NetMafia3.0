let startGame = document.querySelector(".start__game")

startGame.addEventListener("click",joinFastGame)

function joinFastGame(){
    fetch('/joinroombyid', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({roomId : ""})
    }).then(response => {
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