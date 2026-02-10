function getLogin(handle) {
    const url = new URL(window.location.origin + "/handleLogin")
    url.searchParams.append("handle", handle)
    return fetch(url)
}

const handleEntry = document.getElementById("handle");
const atBtn = document.getElementById("atBtn");

if (handleEntry && atBtn) {
    atBtn.addEventListener('click', async function() {
        loginResponse = await getLogin(handleEntry.value);
        if (!loginResponse.ok){
            console.log(loginResponse)
            return
        }
        const data = await loginResponse.json()
        window.location.href = data.redirect;
    });
}
