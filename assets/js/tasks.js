'use strict';

// TODO Error handling

function createTaskContainer(task) {
    $('.tasks-ul').prepend(`
        <li class="task" id="task_${task.uuid}">
            <p class="task-content">${task.value}</p>
            <button class="task-view-button">View</button>
            <button class="task-remove-button">Remove</button>
        </li>
    `);
    $('.task-view-button').on('click', e => {
        viewTask(e.target.parentElement.id.slice(5));
        e.preventDefault();
    });
    $('.task-remove-button').on('click', e => {
        removeTask(e.target.parentElement.id.slice(5));
        e.preventDefault();
    });
}

async function insertTaskRequest(content) {
    let resp = await fetch(
    `http://127.0.0.1:4201/tasks`,
    {
        method: 'POST',
        body: JSON.stringify({
            value: content,
            is_resolved: false
        })
    });
    if (resp.ok) {
        return await resp.json();
    } else {
        throw `Failed to insert task ${resp.status} ${resp.statusText}`
    }
}

function insertTask() {
    let form = $('#tasks-add-form').serializeArray();
    let content = form[0].value;
    insertTaskRequest(content)
        .then((task) => {
            createTaskContainer(task);
            $('#placeholder').remove();
        })
        .catch((err) => {
            console.error(`Failed to insert task`, err)
        });
}

async function viewTaskRequest(id) {
    let resp = await fetch(
        `http://127.0.0.1:4201/tasks/${id}`,
        { method: 'GET' });
    if (resp.ok) {
        return await resp.text();
    } else {
        throw `Failed to view task ${resp.status} ${resp.statusText}`;
    }
}

function viewTask(id) {
    viewTaskRequest()
        .then(result => {
            document.location.href = `http://127.0.0.1:4201/tasks/${id}`
        })
        .catch(err => console.error(`Failed to view task with id: ${id}`, err));
}

async function removeTaskRequest(id) {
    let resp = await fetch(
    `http://127.0.0.1:4201/tasks/${id}`,
    {
        method: 'DELETE'
    });
    if (!resp.ok) {
        throw `Failed to remove task ${resp.status} ${resp.statusText}`
    }
}

function removeTask(id) {
    removeTaskRequest(id)
        .then(() => {
            $(`#task_${id}`).remove();
        })
        .catch((err) => console.error(`Failed to remove task with id: ${id}`, err));
}

// Handlers
$('.tasks-add-form').on('submit', e => {
    insertTask();
    e.preventDefault();
});
$('.task-view-button').on('click', e => {
    viewTask(e.target.parentElement.id.slice(5));
    e.preventDefault();
});
$('.task-remove-button').on('click', e => {
    removeTask(e.target.parentElement.id.slice(5));
    e.preventDefault();
});
