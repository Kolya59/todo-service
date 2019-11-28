'use strict';

async function signInRequest(login, password) {
    let resp = await fetch (
        `http://127.0.0.1:4201/auth/signin`,
        {
            method: 'POST',
            body: JSON.stringify({
                login: login,
                password: password
            }),
            credentials: 'same-origin'
        });
    if (resp.ok) {
        let uuid = await resp.text();
        localStorage.setItem('uuid', uuid);
    } else {
        throw `Failed to sign in ${resp.status} ${resp.statusText}`;
    }
}

function signIn() {
    let form = $('.auth-form').serialize();
    let login = form.login;
    let password = form.password;
    signInRequest(login, password)
        // TODO Put uuid into redirect body
        .then(() => window.location.href = "http://127.0.0.1:4201/tasks")
        .catch((e) => alert(`Failed to sign in ${e}`));
}

async function signUpRequest(login, password) {
    let resp = await fetch (
        `http://127.0.0.1:4201/auth/signup`,
        {
            method: 'POST',
            body: JSON.stringify({
                login: login,
                password: password
            })
        });
    if (resp.ok) {
        let uuid = await resp.text();
        localStorage.setItem('uuid', uuid);
    } else {
        throw `Failed to sign up ${resp.status} ${resp.statusText}`;
    }
}

function signUp() {
    let form = $('.auth-form').serialize();
    let login = form.login;
    let password = form.password;
    signUpRequest(login, password)
        // TODO Put uuid into redirect body
        .then(() => window.location.href = "http://127.0.0.1:4201/tasks")
        .catch((e) => alert(`Failed to sign up ${e}`));
}

// Handlers
$('.auth-form').submit(e => {
    alert('Submit 1');
    e.preventDefault();
});
$('.auth-form-sign-in').submit(e => {
    alert('Submit 2');
    signIn();
    e.preventDefault();
});
$('.auth-form-sign-up').submit(e => {
    alert('Submit 3');
    signUp();
    e.preventDefault();
});