let friendListButton = document.querySelector(".add__friend__menu__tabs__btn-left")
let searchListButton = document.querySelector(".add__friend__menu__tabs__btn-right")
let scrollTable = document.querySelector(".add__friend__menu__scroll")
let friendsList = document.querySelector(".add__friend__list")
let addFriendButton = document.querySelector(".search__button")
let form = document.querySelector(".search__form")
let input = document.getElementById("search__input")
let deleteButtons = []
let popup = document.querySelector(".overflow");
let popupButton = document.querySelector(".button__popup");



document.addEventListener("DOMContentLoaded", fetchFriend);
form.addEventListener("submit", (event) => { addFriend(event)});
popupButton.addEventListener("click", ()=>{popup.classList.remove("active")})



searchListButton.addEventListener("click",()=>{
    scrollTable.classList.add("active")
    searchListButton.classList.add("active")
    friendListButton.classList.remove("active")

})
friendListButton.addEventListener("click",()=>{
    fetchFriend()
    scrollTable.classList.remove("active")
    searchListButton.classList.remove("active")
    friendListButton.classList.add("active")
})

async function fetchFriend() {
    try {
        fetch('/get-list-friends', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
        }).then(response => {
            if (!response.ok) {
                return response.json().then(errData => {
                    throw new Error(errData.error || `Ошибка: ${response.status} ${response.statusText}`);
                });
            }
            return response.json();
        })
        .then(data => {
            friendsList.innerHTML = ''
            if(data.friends.length > 0){
                //добавляем элементы 
                data.friends.forEach(element => {
                    let block = createFriendBlock(element)
                    friendsList.append(block)
                    deleteButtons.push(block.querySelector(".add__friend__friends__block__btn__remove"))
                });
                deleteButtons.forEach((element)=>{
                    element.addEventListener("click", ()=>{removeFriend(element)})
                })
                
            } else {
                friendsList.innerText = "список друзей пуст"
            }
        })
        .catch(error => {
            console.log(error.message)
        });
    } catch (error) {
        console.error("Ошибка запроса:", error);
    }
}
function removeFriend(element){
    let name = element.closest(".add__friend__friends").querySelector(".add__friend__friends__name-font").innerText.trim()
    console.log(name)
    try {
        fetch('/friends-remove', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                friend: name
            })
        }).then(response => {
            if (!response.ok) {
                return response.json().then(errData => {
                    throw new Error(errData.error || `Ошибка: ${response.status} ${response.statusText}`);
                });
            }
            return response.json();
        })
        .then(data => {
            fetchFriend()
        })
        .catch(error => {
            console.log(error.message)
        });
    } catch (error) {
        console.error("Ошибка запроса:", error);
    }

}
async function addFriend(event){
    event.preventDefault(); 

    const friendName = input.value.trim(); 
    if (!friendName) return; 

    fetch('/friends-add', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            friend: friendName
        })
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
            popup.classList.add("active")
            popupButton.innerText = "Продолжить"
            popup.querySelector(".popup__text").innerText = data.message
        })
        .catch(error => {
            console.log(error.message)
        });
}


function createFriendBlock(name) {
    const friendBlock = document.createElement('div');
    friendBlock.classList.add('add__friend__friends');
  
    const nameContainer = document.createElement('div');
    nameContainer.classList.add('add__friend__friends__name');
    nameContainer.textContent = 'name: ';
  
    const nameSpan = document.createElement('span');
    nameSpan.classList.add('add__friend__friends__name-font');
    nameSpan.textContent = name;
    nameContainer.appendChild(nameSpan);
  
    const buttonBlock = document.createElement('div');
    buttonBlock.classList.add('add__friend__friends__block');
  
    // const addButton = document.createElement('button');
    // addButton.classList.add('add__friend__friends__block__btn', 'add__friend__friends__block__btn__add');
    // addButton.textContent = 'Добавить';
  
    const removeButton = document.createElement('button');
    removeButton.classList.add('add__friend__friends__block__btn', 'add__friend__friends__block__btn__remove');
    removeButton.textContent = 'Удалить';
  
    buttonBlock.appendChild(removeButton);
  
    friendBlock.appendChild(nameContainer);
    friendBlock.appendChild(buttonBlock);
  
    return friendBlock;
  }
  
