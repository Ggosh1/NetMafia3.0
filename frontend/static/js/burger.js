let burger = document.querySelector(".burger");
burger.addEventListener("click",burderMenu)

function burderMenu(){
    let menu = document.querySelector(".menu")
    if(menu.classList.contains("active")){
        document.querySelector("body").style.position = "unset"
        menu.classList.remove("active")
        burger.classList.remove("active")
        document.querySelector(".overflow1").style.display = "none"
    } else{
        menu.classList.add("active")
        burger.classList.add("active")
        document.querySelector("body").style.position = "fixed"
        document.querySelector(".overflow1").style.display = "block"
    }
}