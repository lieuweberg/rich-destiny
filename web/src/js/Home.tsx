import React, { useEffect } from "react";
import { Link } from "react-router-dom";

import PresenceCard, { PresenceCardProps } from "./components/PresenceCard";

import "../css/home.scss";

// @ts-expect-error
import hero from "../../static/images/s20-hero.webp"

// @ts-expect-error
import home1 from "../../static/images/home1.webp";
// @ts-expect-error
import home2 from "../../static/images/home2.webp";
// @ts-expect-error
import home3 from "../../static/images/home3.webp";

// @ts-expect-error
import homeWaves1 from "../../static/images/home-waves1.svg";

export default function() {
    let presences: PresenceCardProps[] = [];
    for (let i = 0; i < 3*6; i++) {
        let p = examplePresences[i];
        presences.push({
            largeImage: p.i,
            description: p.d,
            state: p.s,
            initialTime: p.t
        })
    }

    function mouseMove(e: MouseEvent) {
        let hero = document.querySelector("#hero>img");
        let x = (window.innerWidth - e.pageX) / 150;
        let y = (window.innerHeight - e.pageY) / 150;
        // @ts-expect-error
        hero.style.transform = `translateX(${x}px) translateY(${y}px)`;
    }

    useEffect(() => {
        document.addEventListener("mousemove", mouseMove);

        return () => {
            document.removeEventListener("mousemove", mouseMove);
        }
    }, [])
    

    return <>
        <div id="hero">
            <img src={hero} alt="" />
        </div>

        <div id="title">
            <h1>rich-destiny</h1>
            <p>a discord rich presence tool for destiny 2 (pc)</p>
            <div>
                <Link to="/download">Download</Link>
                <Link to="/cp">Control Panel</Link>
            </div>
        </div>
        <div id="sliders">
            <div className="animated-row-1">
                <div className="slider-item">
                    {presences.slice(0,6).map((p, i) => <PresenceCard key={i} {...p} />)}
                </div>
                <div className="slider-item">
                    {presences.slice(0,6).map((p, i) => <PresenceCard key={i} {...p} />)}
                </div>
            </div>
            <div className="animated-row-2">
                <div className="slider-item">
                    {presences.slice(6,12).map((p, i) => <PresenceCard key={i} {...p} />)}
                </div>
                <div className="slider-item">
                    {presences.slice(6,12).map((p, i) => <PresenceCard key={i} {...p} />)}
                </div>
            </div>
            <div className="animated-row-3">
                <div className="slider-item">
                    {presences.slice(12).map((p, i) => <PresenceCard key={i} {...p} />)}
                </div>
                <div className="slider-item">
                    {presences.slice(12).map((p, i) => <PresenceCard key={i} {...p} />)}
                </div>
            </div>
        </div>

        <img className="svg-spacer" src={homeWaves1} alt="" />

        <div id="info-cards">
            <div className="card">
                <div className="image">
                    <img width="300px" height="150px" src={home1} alt="screenshot showing tooltip" />
                </div>
                <hr />
                <div>
                    <h3>Forge your fireteam</h3>
                    <p>Anyone playing Destiny 2 on Steam can easily launch the game through your status. <small>(optional)</small></p>
                </div>
            </div>
            {/* <div className="card">
                <div className="image">
                    <img src={home1} alt="screenshot showing join game button"/>
                </div>
                <hr/>
                <div>
                    <h3>Forge your fireteam</h3>
                    <p>No more sharing join codes: have anyone effortlessly join your fireteam by clicking a button in
                        your status. <small>(disabled)</small></p>
                </div>
            </div> */}
            <div className="card">
                <div className="image">
                    <img width="300px" height="150px" src={home2} alt="screenshot showing tooltip"/>
                </div>
                <hr/>
                <div>
                    <h3>Make your class proud</h3>
                    <p>Displays current class and power level as a tooltip on the class icon.</p>
                </div>
            </div>
            <div className="card">
                <div className="image">
                    <img width="300px" height="150px" src={home3} title="credit: tiberaus" alt="image of errors before arrow, checkmarks after"/>
                </div>
                <hr/>
                <div>
                    <h3>It just works</h3>
                    <p>Runs in the background with minimal resource overhead. Auto updates by default. Fully self-contained.</p>
                </div>
            </div>
        </div>

        <img className="svg-spacer flip" src={homeWaves1} alt="" />

        <div className="generic-text top-padding">
            <h1>What's so special about this?</h1>
            <p>After <Link to="/download">downloading</Link>, this program will
            be started in the background, then sit idly, waiting until you start Destiny 2.</p>
            <p>All alternatives have a built-in User Interface. This adds a lot of
            overhead while rarely being used. Instead, rich-destiny uses a hosted web-based interface for customising
            settings to keep resource usage, installation size and update size minimal.</p>
            <p>Additionally, the intention is that you configure the program once and then never touch it again.</p>
        </div>
    </>
}

const examplePresences = [
    {i:"wellspring",d:"The Wellspring - Savathûn's Throne World",s:"Defend: Master",t:"08:42"},
    {i:"destinylogo",d:"In Orbit",s:"space 🌌 (customisable)",t:"04:13"},
    {i:"beyondlight",d:"Story - Europa",s:"The New Kell",t:"13:24"},
    {i:"raid",d:"Root of Nightmares - Essence",s:"Macrocosm (3/4)",t:"02:15:11"},
    {i:"control",d:"Control - The Crucible",s:"Pacifica",t:"05:58"},
    {i:"gambit",d:"Gambit",s:"Emerald Coast",t:"07:26"},
    {i:"crucible",d:"Mayhem - The Crucible",s:"Javelin-4",t:"02:55"},
    {i:"socialall",d:"Social - Earth",s:"Tower",t:"13:17"},
    {i:"anniversary",d:"Dares of Eternity",s:"Difficulty: Normal",t:"09:08"},
    {i:"strikes",d:"Nightfall: The Ordeal - The Cosmodrome",s:"Difficulty: Master",t:"29:33"},
    {i:"dungeon",d:"Dungeon - Mars",s:"Spire of the Watcher: Normal",t:"43:22"},
    {i:"explore",d:"Explore - The Moon",s:"",t:"06:14"},
    {i:"thewitchqueen",d:"Story - Savathûn's Throne World",s:"The Investigation: Legendary",t:"11:50"},
    {i:"ironbanner",d:"Iron Banner - The Crucible",s:"Midtown",t:"03:27"},
    {i:"lostsector",d:"Lost Sector - Europa",s:"Perdition: Master",t:"11:50"},
    {i:"lightfall", d:"Story - Neptune",s:"First Contact: Legendary",t:"08:32"},
    {i:"trialsofosiris",d:"Trials of Osiris - The Crucible",s:"Convergence",t:"04:50"},
    {i:"seasondefiance",d:"Defiant Battleground - EDZ",s:"Difficulty: Legend",t:"07:25"}
];