![GitHub all releases](https://img.shields.io/github/downloads/lieuweberg/rich-destiny/total) ![GitHub release (latest by date)](https://img.shields.io/github/downloads/lieuweberg/rich-destiny/latest/total) ![Time spent coding (15 minute timeout)](https://wakatime.com/badge/user/a637a12a-da87-4487-8b1e-660151dc3e7b/project/94fa2fc6-7e9b-4c74-b6eb-4ce6a09b4cdf.svg)

# rich-destiny
| <img src="https://richdestiny.app/favicon.ico" width="100"> | Plug-and-play background program that puts your current Destiny 2 activity in your Discord status. Modern, no flaky screenshots and tiny in size. |
| :---: | :---: |

## Contributing ✨

If you want to contribute, awesome! For new features, please ask first on the Discord server or make an issue with what you want to make. If you want to fix a bug, just create a PR. Asking is still recommended however, in case "it's not a bug, it's a feature."

## Developing 🛠
Prerequisites:
 - (client) Have a recent Go 1.x version.
 - (client) For windows: have git bash. You can also run the commands in the build script manually but I use the bash script :)
 - (web) Have a recent version of Node.js and npm.

Building:
 - Clone the repo.
  
 - The client:
   - `cd client`
   - Duplicate the `config.go.example` to `config.go` and fill in the values. For redirect uri, you can use `https://richdestiny.app/login` -- also on the Bungie.net developer portal. It's just a redirect to the localhost redirect.
   - `go get`
   - `./build dev`
     - You can use `./build vX.Y.Z` with a valid semver version number, but for development purposes use dev. It automatically disables updates and possibly other things in the future.

 - The website:
   - `cd web`
   - `npm install`
   - `npm run start` for a local development server or `npm run build` for a production build.