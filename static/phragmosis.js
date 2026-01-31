function getLogin(handle) {
    const url = new URL(window.location.origin + "/handleLogin")
    url.searchParams.append("handle", handle)
    return fetch(url)
}
const handleEntry = document.getElementById("handle");
// handleEntry.addEventListener('input', function(event) {
//     // if the handle looks like a URL (abc.xyz) we try to validate it and shade the entry if good
//     const handle = event.target.value;
//     try {
//         handleURL = new URL(handle)
        
//     } catch (err) {
//         // shade red
//     }
// });
const atBtn = document.getElementById("atBtn");
atBtn.addEventListener('click', async function() {
    
    loginResponse = await getLogin(handleEntry.value);
    if (!loginResponse.ok){
        console.log(loginResponse)
        return
    }
    
    const data = await loginResponse.json()
    console.log(data.redirect)
    window.location.href = data.redirect; 

});