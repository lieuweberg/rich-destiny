import React from "react";

import "../../css/components/PresenceCard.scss";

// @ts-expect-error
import images from "../../images/presence/*.webp";

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
                <img width="60px" height="60px" className="large-image" src={images[largeImage]} draggable="false" alt=""/>
                <img width="20px" height="20px" className="small-image" src={images[smallImage]} draggable="false" alt=""/>
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