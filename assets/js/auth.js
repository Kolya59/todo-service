function signIn() {
    let form = $('#auth-form').serialize();
    let login = form.login;
    let password = form.password;
    fetch(
        `/auth/signin`,
        {
            method: 'POST',
            body: JSON.stringify({
                login: login,
                password: password
            })
        })
        .then((result) => {
            if (result.ok) {
                createTaskContainer(result.body.uuid, form.task_content);
            } else {
                alert(`Failed to sign in ${result.error()}`);
            }
        })
        .catch((err) => {
            alert(`Failed to sign in ${err}`);
        });
}

function signUp() {
    let form = $('#auth-form').serialize();
    let login = form.login;
    let password = form.password;
    fetch(
        `/auth/signup`,
        {
            method: 'POST',
            body: JSON.stringify({
                login: login,
                password: password
            })
        })
        .then((result) => {
            if (result.ok) {
                window.location.href = "/tasks";
            } else {
                alert(`Failed to sign up ${result.error()}`);
            }
        })
        .catch((err) => {
            alert(`Failed to sign up ${err}`);
        });
}