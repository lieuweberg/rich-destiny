![GitHub all releases](https://img.shields.io/github/downloads/lieuweberg/rich-destiny/total) ![GitHub release (latest by date)](https://img.shields.io/github/downloads/lieuweberg/rich-destiny/latest/total)

# rich-destiny
| Plug-and-play background program that puts your current Destiny 2 activity in your Discord status. Modern, no flaky screenshots and tiny in size. | <img src="https://richdestiny.app/rich-destiny.ae89fafb.png" width="100"> |
| :---: | :---: |

## Contributing ✨

If you want to contribute, awesome! For new features, please ask first on the Discord server or make an issue with what you want to make. If you want to fix a bug, just create a PR. Asking is still recommended however, in case "it's not a bug, it's a feature."

## Developing 🛠
Prerequisites:
 - (client/backend) Have at least go 1.15.5. I'm not sure what other versions will work, but I use this.
 - (client) For windows: have git bash.
 - (web) Have a recent version of Node.js and npm.

Building:
 - Clone the repo.
  
 - The client:
   - `cd client`
   - `go get`
   - `./build dev`
     - You can use `./build v0.0.0` with a valid semver version number, but for development purposes use dev.

 - The website:
   - `cd web`
   - `npm install`
   - `npm run start` for a local development server or `npm run build` for a production build.