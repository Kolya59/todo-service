'use strict';

function createTaskContainer(task) {
    $('.tasks-ul').prepend(`
        <li class="task" id="task_${task.id}">
            <p class="task-content">${task.content}</p>
            <button class="task-view-button">View</button>
            <button class="task-remove-button">Remove</button>
        </li>
    `);
}

async function insertTaskRequest(content) {
    let resp = await fetch(
    `http://127.0.0.1:4201/tasks`,
    {
        method: 'POST',
        body: JSON.stringify({
            author: localStorage.getItem('uuid'),
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
    let form = $('#tasks-add-form').serialize();
    let content = form.task_content;
    insertTaskRequest(content)
        .then((task) => {
            createTaskContainer(task)
        })
        .catch((err) => {
            console.error(`Failed to insert task`, err)
        });
}

async function viewTaskRequest(id) {
    let resp = await fetch(
        `http://127.0.0.1:4201/tasks/${id}`,
        {
            method: 'GET',
            headers: JSON.stringify({
                Authorization: localStorage.getItem('uuid'),
            })
        });
    if (resp.ok) {
        return await resp.text();
    } else {
        throw `Failed to view task ${resp.status} ${resp.statusText}`;
    }
}

function viewTask(id) {
    viewTaskRequest()
        .then(result => {
            // TODO Refactor
            document.location.href = `http://127.0.0.1:4201/tasks/${id}`
        })
        .catch(err => console.error(`Failed to view task with id: ${id}`, err));
}

async function removeTaskRequest(id) {
    let resp = await fetch(
    `http://127.0.0.1:4201/tasks/${id}`,
    {
        method: 'DELETE',
        headers: JSON.stringify({
            Authorization: localStorage.getItem('uuid'),
        })
    });
    if (resp.ok) {
        $('.tasks-ul').remove(`#task${id}`)
    } else {
        throw `Failed to remove task ${resp.status} ${resp.statusText}`
    }
}

function removeTask(id) {
    removeTaskRequest(id)
        .then(() => {
            $('.tasks-ul').remove(`#task_${id}`);
        })
        .catch((err) => console.error(`Failed to remove task with id: ${id}`, err));
}

// Handlers
$('.tasks-add-form').on('submit', e => {
    e.preventDefault();
    insertTask();
});
$('.task-view-button').on('click', e => {
    e.preventDefault();
    viewTask(e.target.parentElement.id.slice(5));
});
$('.task-remove-button').on('click', e => {
    e.preventDefault();
    removeTask(e.target.parentElement.id.slice(5));
});
