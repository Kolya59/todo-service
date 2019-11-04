'use strict';

function createTaskContainer(id, content) {
    $('.tasks-ul').prepend(`
        <li class="task" id="task_${id}">
            <p class="task-content">${content}</p>
            <button class="task-view-button">View</button>
            <button class="task-remove-button">Remove</button>
        </li>
    `);

    /*// Set elements
    let tasks_ul = document.getElementsByClassName('tasks-ul').item(0);

    let task_li = document.createElement('li');
    task_li.className = 'task';

    let task_content = document.createElement('p');
    task_content.className = 'task-content';
    task_content.value = content;

    let view_task_button = document.createElement('button');
    view_task_button.className = 'task-view-button';
    view_task_button.onclick = _ => viewTask(id);
    view_task_button.value = id;

    let remove_task_button = document.createElement('button');
    remove_task_button.className = 'task-remove-button';
    remove_task_button.onclick = _ => removeTask(id);
    remove_task_button.value = id;


    // Set relationships
    task_li.appendChild(task_content);
    task_li.appendChild(view_task_button);
    task_li.appendChild(remove_task_button);
    tasks_ul.appendChild(task_li);*/
}

function insertTask() {
    let form = $('#tasks-add-form').serialize();
    let content = form.task_content;
    fetch(
        `/tasks`,
        {
            method: 'POST',
            body: JSON.stringify({
                token: '123best',
                content: content
            })
        })
        .then((result) => {
            if (result.ok) {
                createTaskContainer(result.body.uuid, form.task_content);
            } else {
                console.error(`Failed to insert task`, result.body)
            }
        })
        .catch((err) => {
            console.error(`Failed to insert task`, err)
        });
}

function viewTask(id) {
    document.location.href = `/tasks/${id}`
}

function removeTask(id) {
    fetch(
        `/tasks/${id}`,
        {
            method: 'DELETE',
            body: JSON.stringify({token: '123best'})
        })
        .then((result) => {
            if (result.ok) {
                $('.tasks-ul').remove(`#task${id}`)
            } else {
                console.error(`Failed to remove task`, result.body)
            }
        })
        .catch((err) => {
            console.error(`Failed to remove task with id: ${id}`, err)
        });
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
