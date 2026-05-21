let theme = getCookie("theme")
if (theme === undefined) {
    if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
        document.body.classList.remove('light')
        theme = "dark"
    }
    else {
        document.body.classList.add('light')
        theme = "light"
    }
}
else {
    if (theme == "light") {
        document.body.classList.add('light')
        theme = "light"
    }
    else {
        document.body.classList.remove('light')
        theme = "dark"
    }
}

function toggleTheme() {
    if (theme == "light") {
        document.body.classList.remove('light')
        theme = "dark"
    }
    else {
        document.body.classList.add('light')
        theme = "light"
    }
    document.cookie = `theme=${theme}; max-age=2592000; Path=/`;
}

function getCookie(name) {
    let matches = document.cookie.match(new RegExp(
        "(?:^|; )" + name.replace(/([\.$?*|{}\(\)\[\]\\\/\+^])/g, '\\$1') + "=([^;]*)"
    ));
    return matches ? decodeURIComponent(matches[1]) : undefined;
}

