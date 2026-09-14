# Boring Tasks Tracker :)

This is a boring idea but very useful to understand a few concepts i have been trying to learn. It also kind of solves a real "problem" i have myself. I like to use the terminal and Neovim. I have a whole setup to enjoy writing tasks, notes and code there. If i have my computer available, i prefer to write my daily todo lists there since it looks cool and keeps me in the same setup i already enjoy and know.

Whenever I am away from the computer though, I don't have access to this setup, and this breaks the whole workflow. I want to combine both environments by being able to update the lists from different machines and from anywhere, without relying on other type of tools like Apple Notes or another type of app that can stay in sync.

Even though there are more tools that i could just use straight away, I find this to be a very nice way to understand concepts such as deployment, avilability, networking, server-client architectures, and specially, as practice to learn [fastAPI](https://fastapi.tiangolo.com), [sqlmodel](https://sqlmodel.tiangolo.com), [SQLite](https://sqlite.org/) and a little bit of frontend.

---

# How does it work?

- The source of truth for the tasks is an SQLite database. I have written a fastAPI backend to handle all interactions with the database, and that takes care of most of the functionality such as sorting, cascading effects for completion and deletion and so on. After that is in place i will host that somewhere and there would be two "dumb" clients that will send requests to it.
- Initially the idea was to base everything from Neovim, but i have moved on into building a TUI client written in Go, using the [Bubbletea](https://github.com/charmbracelet/bubbletea) framework. That takes care of the terminal side of things and eventually I will look into serving that frontend through ssh, so that i can actually use this everywhere. For now the go code or binary needs to live in my machine.
- Then, there will be a web frontend that emulates the same functionality as the TUI, so that this becomes available from my phone, or any other device with a browser and internet access.
- The last part would be to self host both the backend and database, as well as the frontend, so it becomes available anywhere, but its also free. Tailscale and an old linux laptop at home will probably be the way to do this.
- Ideally, the end product will consist of providing a way to clone this repo, run a few commands (something like a Makefile in combination with Docker), and the whole network is ready for anyone to use themselves. That will take some time though.

---

# AI use:

I have gone the path of using a lot of AI generated code before and that has yielded only broken programs that i don't understand. I have intentionally chosen to write every line here by hand, using an AI agent when i need to ask about bugs, but not to generate any of the code i care about. I have also used it to test an idea, like creating the web and TUI clients just to test the concept and put the backend to test straight away, however now that i see they work, i will wipe that and build them by hand as well.

I have even chosen to avoid the use of AI generated text in the documents like this README or the WORKING_PLAN i created, as an exercise to double check my understanding of the project, the interactions between the different parts and just to write something that will hopefully else be understood by another human.

I am not against the use of AI, however from my little experience with it i have seen that at the stage i am in my learning journey, it actually becomes an obstacle to my skill development. I am finding ways to use it as a multiplier of my own capacity rather than as the engine that thinks and executes for me.
