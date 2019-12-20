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

function changeStatus(id, oldValue) {
    let newValue = !oldValue;
    changeStatusRequest(id, oldValue)
        .then(() => {
            $('.is_resolved').attr('checked', `${newValue}`);
        })
        .catch((e) => {
            alert(`Failed to switch task status ${e}`);
            $('.is_resolved').attr('checked', `${oldValue}`);
        })
}

$('.is_resolved').on('click', e => {
   $('#form').submit();
});

$('#form').on('submit', e => {
    let form = $('#form').serializeArray();
    let old_value = form[1] ? form[1].value === 'on' : false;
    changeStatus(form[0].value, old_value);
    e.preventDefault();
});