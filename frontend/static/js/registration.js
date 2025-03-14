window.addEventListener("load", ()=>{
    let loginButton = document.querySelector("#login");
    let registerButton = document.querySelector("#register");
    let formsBlock = document.querySelector(".forms__block");
    let registration = document.querySelector(".registration");
    let allPaswwrodsInput = document.querySelectorAll("[data-password]");
    let password = "";
    let allLoginsInput = document.querySelectorAll("[data-login]")
    let loginButtonSubmit = document.querySelector("[data-log]");
    let RegistrButtonSubmit = document.querySelector("[data-regist]");
    let popup = document.querySelector(".overflow");
    let popupButton = document.querySelector(".button__popup");
    loginButton.addEventListener("click", ()=>{
        if(!loginButton.classList.contains("active")){
            loginButton.classList.add("active");
            registerButton.classList.remove("active");
            formsBlock.classList.remove("active")
            registration.classList.remove("active");
        }
    })
    registerButton.addEventListener("click", ()=>{
        if(!registerButton.classList.contains("active")){
            loginButton.classList.remove("active");
            registerButton.classList.add("active");
            formsBlock.classList.add("active")
            registration.classList.add("active");
        }
    })
    allPaswwrodsInput.forEach((element)=>{
        element.addEventListener("input",()=>{
            password = element.value
            element.value = "*".repeat(password.length) 
            if(password.length < 8){
                element.classList.remove("allowed")
                element.closest(".block__inputs").querySelector(".warning__message").innerText= "Длина пароля должна быть больше 8 символов"
                element.style.borderColor = "red";
                element.style.borderWidth = "2px";
            }
            if(password.length >= 8){
                element.classList.add("allowed")
                element.style.borderColor = "#03182c";
                element.style.borderWidth = "1px";
                element.closest(".block__inputs").querySelector(".warning__message").innerText= ""
            }
        })
    })
    allLoginsInput.forEach((element) => {
        element.addEventListener("input", ()=>{
            if(element.value == ""){
                element.classList.remove("allowed")
                element.style.borderColor = "red";
                element.style.borderWidth = "2px";
            } else{
                element.classList.add("allowed")
                element.style.borderColor = "#03182c";
                element.style.borderWidth = "1px";
            }
        })
    })
    RegistrButtonSubmit.addEventListener("click",(e)=>{
        
        let passwordInput = document.querySelector("#registBlock").querySelector("[data-password]")
        let loginInput = document.querySelector("#registBlock").querySelector("[data-login]")
        e.preventDefault();
        if(passwordInput.classList.contains("allowed") && loginInput.classList.contains("allowed")){
            let password = passwordInput.value
            let username = loginInput.value.trim()
            passwordInput.value = ""
            loginInput.value = ""
            fetch('/register', {
                  method: 'POST',
                  headers: {
                    'Content-Type': 'application/json'
                  },
                  body: JSON.stringify({ username: username, password: password })
            })
            .then(response => {
                console.log("Ответ от /register:", response.status);
                if (!response.ok) {
                    return response.text().then(text => { throw new Error(text); });
                }
                return response.json();
            })
            .then(data => {
                popup.classList.add("active")
                popupButton.innerText = "Продолжить"
                popup.querySelector(".popup__text").innerText = data.message
                let span = document.createElement("span")
                span.classList.add("mafia__text")
                span.textContent = data.addMessage
                popup.querySelector(".popup__text").append(span)
                popupButton.addEventListener("click", ()=>{
                    window.location.href = '/profile';
                })
            })
            .catch(error => {
                    popup.classList.add("active")
                    popupButton.innerText = "Попробовать еще раз"
                    popup.querySelector(".popup__text").innerText = error.message
                    popupButton.addEventListener("click", ()=>{
                    popup.classList.remove("active")
                })
            });
        }
        
    })
    loginButtonSubmit.addEventListener("click",(e)=>{
        e.preventDefault();
        let passwordInput = document.querySelector("#loginBlock").querySelector("[data-password]")
        let loginInput = document.querySelector("#loginBlock").querySelector("[data-login]")
        if(passwordInput.classList.contains("allowed") && loginInput.classList.contains("allowed")){
            let password = passwordInput.value
            let username = loginInput.value.trim()
            passwordInput.value = ""
            loginInput.value = ""
            fetch('/login', {
                method: 'POST',
                headers: {
                'Content-Type': 'application/json'
                },
                body: JSON.stringify({ username: username, password: password })
            })
            .then(response => {
                if (!response.ok) {
                return response.json().then(data => { throw new Error(data.error); });
                }
                return response.json();
            })
            .then(data => {
                popup.classList.add("active")
                popupButton.innerText = "Продолжить"
                popup.querySelector(".popup__text").innerText = data.message
                let span = document.createElement("span")
                span.classList.add("mafia__text")
                span.textContent = data.addMessage
                popup.querySelector(".popup__text").append(span)
                popupButton.addEventListener("click", ()=>{
                    window.location.href = '/profile';
                })
            })
            .catch(error => {
                popup.classList.add("active")
                popupButton.innerText = "Попробовать еще раз"
                popup.querySelector(".popup__text").innerText = error.message
                popupButton.addEventListener("click", ()=>{
                    popup.classList.remove("active")
                })
            });
        }
    })
})