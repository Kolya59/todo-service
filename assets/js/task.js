'use strict';

async function changeStatusRequest(id, new_value) {
    let resp = await fetch(
    `http://127.0.0.1:4201/tasks/${id}`,
    {
        method: 'PUT',
        body: JSON.stringify({
            is_resolved: new_value
        })
    });
    if (resp.ok) {
        $(`#${id}`).checked = new_value;
    } else {
        throw `Failed to switch status ${resp.status} ${resp.statusText}`;
    }
}

function changeStatus(id, old_value) {
    let new_value = !old_value;
    changeStatusRequest(id, new_value)
        .then(() => {
            $('.is_resolved').replaceWith(`<input type="checkbox" checked="${new_value}">`);
        })
        .catch((e) => {
            alert(`Failed to switch task status ${e}`);
            $('.is_resolved').replaceWith(`<input type="checkbox" checked="${old_value}">`);
        })
}

$('#form').on('submit', e => {
    let form = $('#form').serializeArray();
    changeStatus(form[0].value, form[1].value);
    e.preventDefault();
});