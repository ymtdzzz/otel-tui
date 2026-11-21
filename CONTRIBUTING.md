# Contributing to otel-tui

## Before you start

Pick the path that matches your change.

| Your change | What to do |
| --- | --- |
| A small fix with no issue, such as a typo, a docs fix, or an obvious one-line bug | Open a pull request directly. |
| Work on an existing issue | Comment on the issue first and wait until a maintainer assigns it to you. |
| A new feature or a larger change with no issue | Open an issue to discuss it first, then follow the row above. |

### Working on an existing issue

1. Comment on the issue to say you'd like to work on it. Include a short note on your approach.
2. Wait for a maintainer to assign the issue to you. Don't open a ready-for-review pull request before that.
3. After you are assigned, implement the change and open a pull request that includes `Closes #<issue number>`.

You can link a draft pull request in your comment if you already have a prototype. See [draft pull requests](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-pull-requests#draft-pull-requests). Mark it ready for review after you are assigned.

Don't start on an issue that is assigned to someone else. If the issue has had no activity for a while, ask on the issue. Maintainers unassign an issue after two weeks with no activity so someone else can work on it.

Maintainers may close a pull request for an existing issue if its author skipped these steps. The steps stop two contributors from writing the same fix. They also let maintainers agree on the approach before you write the code.

## Creating a pull request

1. Fork this repository to your GitHub account.
2. Create a new branch and implement your changes on it.
3. Add tests that cover your changes.
4. Run the linter and the test suites. All of them must pass.

   ```
   make lint
   make test
   make test-exporter
   ```

5. Push the branch to your fork.
6. Open a pull request against `main` and fill in the pull request template.

## AI-assisted contributions

You can use AI tools to write the code and the pull request description. You are still responsible for the change. Before you mark the pull request ready for review, read the whole diff and make sure you understand it. Be ready to answer review comments about it yourself.
