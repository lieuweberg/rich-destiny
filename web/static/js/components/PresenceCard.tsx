import React from "react";

import "../../css/components/PresenceCard.scss";

export type PresenceCardProps = {
    description:    string
    state?:         string
    largeImage:     string
    smallImage?:    string
    initialTime?:   string
}

export default function({largeImage, description, initialTime, state, smallImage}: PresenceCardProps) {
    const [timeMs, setTime] = React.useState(0)
    let time = new Date(timeMs);
    React.useEffect(() => {
        if (initialTime && initialTime.includes("T")) {
            let elapsed = Date.now() - (new Date(initialTime)).getTime();
            setTime(elapsed);
            time = new Date(elapsed);
        } else {
            if (!initialTime) {
                initialTime = "12:34"
            }
            let times = initialTime.split(":");
            time.setSeconds(parseInt(times.pop()));
            time.setMinutes(parseInt(times.pop()));
            if (times.length == 1) {
                time.setHours(parseInt(times[0]));
            }
            setTime(time.getTime());
        }

        let interval = setInterval(() => {
            time.setSeconds(time.getUTCSeconds() + 1);
            setTime(time.getTime());
        }, 1000)
        return () => clearInterval(interval);
    }, [initialTime])

    let t = fmtTime(time.getUTCMinutes()) + ":" + fmtTime(time.getUTCSeconds());
    if (time.getUTCHours() != 0) {
        t = fmtTime(time.getUTCHours()) + ":" + t;
    }

    if (!smallImage) {
        [smallImage] = React.useState(
            ["hunter", "warlock", "titan"][Math.floor(Math.random() * 3)]
        );
    }

    return <div className="presence-wrapper">
        <div className="type">
            <p>playing a game</p>
        </div>
        <div className="presence">
            <div className="images">
                <img className="large-image" src={"https://cdn.discordapp.com/app-assets/726090012877258762/"
                    + imageIdMap[largeImage] + ".png"} draggable="false" alt=""/>
                <img className="small-image" src={"https://cdn.discordapp.com/app-assets/726090012877258762/"
                    + imageIdMap[smallImage] + ".png"} draggable="false" alt=""/>
            </div>
            <div className="text">
                <p className="game">Destiny 2</p>
                <p title={description}>{description}</p>
                <p title={state}>{state}</p>
                <p>{t} elapsed</p>
                {!state ? <p>&nbsp;</p> : ""}
            </div>
        </div>
    </div>
}

function fmtTime(t: number): string {
    if (t < 10) return "0" + t;
    return t.toString();
}

const imageIdMap = {
    control: "726487437026656398",
    crucible: "726487437744013322",
    destinylogo: "726090605373161523",
    doubles: "726487438788395101",
    dungeon: "726487437265862688",
    explore: "726487438419165248",
    forge: "726487439010431126",
    gambit: "726487439211888731",
    hauntedforest: "763440501092384820",
    hunter: "726487437572046860",
    ironbanner: "726487439325003887",
    lostsector: "838469802300801065",
    menagerie: "726487439048441857",
    nightmarehunt: "726487439060762664",
    privatecrucible: "726487438834532353",
    raid: "726487439408889896",
    reckoning: "726487439073607762",
    socialall: "726487439102967890",
    storypvecoopheroic: "726487439438381086",
    strikes: "726487439157362708",
    thenine: "726487438402256967",
    titan: "726487437492224011",
    trialsofosiris: "726487439220408360",
    vexoffensive: "726487438062780497",
    warlock: "726487437978763304"
}