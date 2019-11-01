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
            window.document.write(result);
        })
        .catch((err) => {
            console.error(`Failed to remove task with id: ${id}`, err)
        });
}