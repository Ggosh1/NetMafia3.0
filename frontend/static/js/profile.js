async function fetchLogin() {
    try {
        let response = await fetch("/get-login", { method: "GET", credentials: "include" });
        let data = await response.json();
        
        if (data.login) {
            document.querySelector("[data-login]").textContent = data.login;
        }
    } catch (error) {
        console.error("Ошибка запроса:", error);
    }
}

document.addEventListener("DOMContentLoaded", fetchLogin);