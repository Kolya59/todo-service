'use strict';

async function changeStatusRequest(id, new_value) {
    let resp = await fetch(
    `http://127.0.0.1:4201/tasks/${id}`,
    {
        method: 'PUT',
        body: JSON.stringify({
            uuid: id,
            is_resolved: new_value
        })
    });
    if (resp.ok) {
        return;
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

$('.is_resolved').on('click', e => {
    e.preventDefault();
    changeStatus(e.target.id, e.target.checked());
});