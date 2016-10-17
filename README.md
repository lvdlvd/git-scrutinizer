# git-scrutinizer
A code review tool for git repositories.

THIS IS AT THE CRUDE PROTOTYPE THROWN TOGETHER IN A FEW HOLIDAY HOURS STAGE

Git scrutinizer allows collaborators to add comments per file:line on the differences between master..HEAD in the repository in which it is started.

Unlike other things out there it runs locally (it opens a browser to a localhost:port for the UI) and stores the review threads as structured text messages in git notes instead of in a separate database.

This means it re-uses the authentication, authorisation, communication and storage facilities git already provides and avoids installation struggles.

The only non-go dependency is libgit2 (through the git2go module).

INSTALLATION
- install libgit2 through whatever native means your platform uses
- go get -u github.com/lvdlvd/git-scrutinizer

TODO:
- ui sucks, rethink
- currently can only review master..HEAD
- automate git push origin refs/notes/scrutinize/...  and  git fetch origin refs/notes/scrutinize/*:refs/notes/scrutinize/...
- reply-to messages
- better diff and tree viewers
