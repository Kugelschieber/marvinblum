console.log("Hi! Welcome to my website.");

document.addEventListener("DOMContentLoaded", () => {
    const nav = document.querySelector("#nav");

    document.querySelector(".burger").addEventListener("click", () => {
        if (nav.classList.contains("open")) {
            nav.classList.remove("open");
        } else {
            nav.classList.add("open");
        }
    });
});
