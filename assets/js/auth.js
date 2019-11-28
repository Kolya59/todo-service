'use strict';

function handle() {
    let form = $('.auth-form').serializeArray();
    switch (form[2].value) {
        case 'signin':
            signIn(form);
            break;
        case 'signup':
            signUp(form);
            break;
        default:
            return;
    }
}

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
    if (!resp.ok) {
        throw `Failed to sign in ${resp.status} ${resp.statusText}`;
    }
}

function signIn(form) {
    let login = form[0].value;
    let password = form[1].value;
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
    if (!resp.ok) {
        throw `Failed to sign up ${resp.status} ${resp.statusText}`;
    }
}

function signUp(form) {
    let login = form[0].value;
    let password = form[1].value;
    signUpRequest(login, password)
        // TODO Put uuid into redirect body
        .then(() => window.location.href = "http://127.0.0.1:4201/tasks")
        .catch((e) => alert(`Failed to sign up ${e}`));
}

// Handlers
$('.auth-form').submit(e => {
    handle();
    e.preventDefault();
});
$('.auth-form-sign-in').click(e => {
    $('#hidden').val('signin');
    $('.auth-form').submit();
});
$('.auth-form-sign-up').click(e => {
    $('#hidden').val('signup');
    $('.auth-form').submit();
});