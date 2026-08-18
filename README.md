# Tasks Tracker

This is a boring idea but very useful to understand a few concepts i have been trying to learn. It also kind of solves a real "problem" i have myself. I like to use the terminal and Neovim. I have a whole setup to enjoy writing tasks, notes and code there. If i have my computer available, i prefer to write my daily todo lists there since it looks cool and keeps me in the same setup i already enjoy and know.

Whenever I am away from the computer though, I don't have access to this setup, and this breaks the whole workflow. I want to combine both environments by being able to update the lists from different machines and from anywhere, without relying on other type of tools like Apple Notes or another type of app that can stay in sync.

Even though there are more tools that i could just use straight away, I find this to be a very nice way to understand concepts such as deployment, avilability, networking, server-client architectures, and specially, as practice to learn [fastAPI](https://fastapi.tiangolo.com), [sqlmodel](https://sqlmodel.tiangolo.com), [SQLite](https://sqlite.org/) and a little bit of frontend.

---

# How does it work?

- Basically, the notes are written in a `markdown` file with Neovim.
- On a write event, I will trigger an action to parse the markdown, and identify the individual tasks. That will then check the database to add new tasks, update existing ones or delete the ones that have been removed from the file. The completion status can also be updated by ticking the boxes.
- On the other side, by visiting the corresponding URL, i could see the same list of tasks in my phone or any other device via a web interphase. I can toggle the task's completion status, add new ones or modify existing ones. This should in turn also update the same database to maintain consistency.
- Last, i should also be able to somehow refresh my local file when I am working in my main machine, and the markdown should get generated from the database, using it as the only source of truth.
