# CLAUDE.md

Project rules. These extend the guidelines in `c:\Projects\CLAUDE.md`.

## Writing

1. No em dashes. Use a comma, colon, or period instead.
2. No long comments. Add a short one-line comment only when the code is not self-explanatory. No block comments, no docstring-style headers, no commentary restating what the code does.

## Documentation

3. This is a public open source project. Every document generated here (README, guides, plans, specs, issue and PR text) is written for an outside reader with no prior context.
   - Lead with what the thing is and why it exists, before any detail.
   - Use clear headings, short paragraphs, and lists over walls of text.
   - Keep one topic per section and order sections so a reader can follow start to finish.
   - Define project-specific terms on first use. No unexplained internal shorthand.
   - Prefer plain language and concrete examples over abstract description.
   - Keep it current. Update the affected docs in the same change, do not leave stale instructions behind.

## Commits

4. Commit messages: one short line, then a description.
   - The subject line is a plain sentence saying what the commit does. No `fix:`, `feat:`, `chore:` or any other prefix tag. No trailing period.
   - Leave a blank line, then a description covering three things, in this order:
     - **Why:** what this commit is for, the problem or need behind it.
     - **What changed:** the actual changes, briefly. Not a file list.
     - **To test:** what someone should check to confirm it works.
   - Keep each part to a sentence or two. If a section has nothing worth saying, say so in a few words rather than padding it.

Example:

```
Resume model downloads instead of restarting them

Why: the large model is 3GB and users on unreliable connections were
losing the whole download on a dropped connection.

What changed: downloads write to a .part file and reopen with an HTTP
Range request. SHA256 verification now runs before the file is moved
into place, so a partial file can never be used as a model.

To test: start a large model download, kill the network mid transfer,
reconnect, and confirm it resumes from where it stopped rather than
starting at zero.
```
