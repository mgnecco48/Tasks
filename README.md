# Boring Tasks Tracker :)

This is a boring idea but very useful to understand a few concepts i have been trying to learn. It also kind of solves a real "problem" i have myself. I like to use the terminal and Neovim. I have a whole setup to enjoy writing tasks, notes and code there. If i have my computer available, i prefer to write my daily todo lists there since it looks cool and keeps me in the same setup i already enjoy and know.

Whenever I am away from the computer though, I don't have access to this setup, and this breaks the whole workflow. I want to combine both environments by being able to update the lists from different machines and from anywhere, without relying on other type of tools like Apple Notes or another type of app that can stay in sync.

Even though there are more tools that i could just use straight away, I find this to be a very nice way to understand concepts such as deployment, avilability, networking, server-client architectures, and specially, as practice to learn [fastAPI](https://fastapi.tiangolo.com), [sqlmodel](https://sqlmodel.tiangolo.com), [SQLite](https://sqlite.org/) and a little bit of frontend.

---

# How does it work?

- Basically, the notes are written in a `markdown` file with Neovim.

> [!NOTE]
> Right now, this idea adds a lot of complexity, as i will also have to manage the state of the local file and write a Neovim plugin. I decided to create a TUI that can act as a client to the backend instead to start testing and iterating

- On a write event, I will trigger an action to parse the markdown, and identify the individual tasks. That will then check the database to add new tasks, update existing ones or delete the ones that have been removed from the file. The completion status can also be updated by ticking the boxes.
- On the other side, by visiting the corresponding URL, i could see the same list of tasks in my phone or any other device via a web interphase. I can toggle the task's completion status, add new ones or modify existing ones. This should in turn also update the same database to maintain consistency.
- Last, i should also be able to somehow refresh my local file when I am working in my main machine, and the markdown should get generated from the database, using it as the only source of truth.

---

# AI use:

I have gone the path of using a lot of AI generated code before and that has yielded only broken programs that i don't understand. I have intentionally chosen to write every line here by hand, using an AI agent when i need to ask about bugs, but not to generate any of the code i care about. I have also used it to test an idea, like creating the web and TUI clients just to test the concept and put the backend to test straight away, however now that i see they work, i will wipe that and build them by hand as well.

I have even chosen to avoid the use of AI generated text in the documents like this README or the WORKING_PLAN i created, as an exercise to double check my understanding of the project, the interactions between the different parts and just to write something that will hopefully else be understood by another human.

I am not against the use of AI, however from my little experience with it i have seen that at the stage i am in my learning journey, it actually becomes an obstacle to my skill development. I am finding ways to use it as a multiplier of my own capacity rather than as the engine that thinks and executes for me.
