/** @param {NS} ns */
export async function main(ns) {
    ns.disableLog("ALL");
    ns.clearLog();
    ns.tail();
    const { opponent, size: boardSize } = ns.flags([["opponent", "Netburners"], ["size", 5]]);

    while (true) {
        let gameOver = false;
        const { boardX, boardY } = await findRectBoard(ns, opponent, boardSize);
        const offsetY = boardSize - (boardY + 1);
        const handicaps = findHandicaps(ns, offsetY);
        const isReady = await checkReady(ns);
        if (!isReady) {
            ns.print("ERROR", "SERVER IS NOT READY!");
            return;
        }
        await initBoard(ns, boardX, boardY, handicaps);
        await ns.sleep(5000);

        // Game loop
        while (!gameOver) {
            const currentPlayer = ns.go.getCurrentPlayer();
            if (currentPlayer === "None") {
                ns.clearLog();
                gameOver = true;
                break;
            }

            const bestMove = await genMove(ns, offsetY);
            if (bestMove) {
                await ns.go.makeMove(bestMove.x, bestMove.y);
            } else {
                ns.print("No valid moves, passing turn...");
                await ns.go.passTurn();
            }

            const opMove = await ns.go.opponentNextTurn()

            if (opMove.type === "move") {
                await playMove(ns, numbersToLetters[opMove.x] + (opMove.y - offsetY));
            } else if (opMove.type === "pass") {
                await playMove(ns, opMove.type);
            } else {
                ns.clearLog();
                gameOver = true;
                break;
            }
            await ns.sleep(500);
        }
    }
}

/** @param {NS} ns */
const checkReady = async (ns) => {
    for (let loopCount = 0; loopCount < 60; loopCount++) {
        const res = await fetch("http://localhost:8080/check-ready", {
            method: "GET",
            headers: {
                "Accept": "*/*",
                "Authorization": "...",
                "Content-Type": "application/json",
                "Connection": "keep-alive"
            }
        });

        if (res.status === 200) {
            return true;
        }

        await ns.sleep(1000);
    }
    return false;
}

/**
 * @param {NS} ns
 * @param {number} boardX
 * @param {number} boardY
 * @param {string[]} handicaps
 */
const initBoard = async (ns, boardX, boardY, handicaps) => {
    const komi = ns.go.getGameState().komi;
    const reqBody = { "board-size": `${boardX} ${boardY}`, "komi": `${komi}`, "handicaps": handicaps };
    ns.print("INIT-BOARD REQUEST:", reqBody);
    await fetch("http://localhost:8080/init", {
        method: "POST",
        headers: {
            "Accept": "*/*",
            "Authorization": "...",
            "Content-Type": "application/json",
            "Connection": "keep-alive"
        },
        body: JSON.stringify(reqBody)
    });
}

/**
 * @param {NS} ns
 * @param {string} moveToPos
 */
const playMove = async (ns, moveToPos) => {
    ns.print("PLAY-MOVE REQUEST:", moveToPos);
    await fetch("http://localhost:8080/play-move", {
        method: "POST",
        headers: {
            "Accept": "*/*",
            "Authorization": "...",
            "Content-Type": "application/json",
            "Connection": "keep-alive"
        },
        body: JSON.stringify({ color: "white", "move_to_pos": moveToPos })
    });
}

/**
 * @param {NS} ns
 * @param {number} offsetY
 */
const genMove = async (ns, offsetY) => {
    const res = await fetch("http://localhost:8080/gen-move", {
        method: "POST",
        headers: {
            "Accept": "*/*",
            "Authorization": "...",
            "Content-Type": "application/json",
            "Connection": "keep-alive"
        },
        body: JSON.stringify({ color: "black" })
    });
    if (res.status === 200) {
        const data = await res.json();
        const dataMove = data.move;
        ns.print("GEN-MOVE RESPONSE:", dataMove);
        if (dataMove && dataMove !== "pass") {
            return {
                x: lettersToNumbers[dataMove.at(0)],
                y: Number(dataMove.substring(1)) + offsetY
            };
        }
    }
    return null;
}

/**
 * @param {NS} ns
 * @param {number} offsetY
 */
const findHandicaps = (ns, offsetY) => {
    const matrix = ns.go.getBoardState().map(row => row.split(''));
    let handicaps = [];
    for (let i = 0; i < matrix.length; i++) {
        for (let j = 0; j < matrix[i].length; j++) {
            if (matrix[i][j] === "O") {
                handicaps.push(numbersToLetters[i] + (j - offsetY));
            }
        }
    }
    return handicaps;
}

/**
 * @param {NS} ns
 * @param {string} opponent
 * @param {number} boardSize
 */
const findRectBoard = async (ns, opponent, boardSize) => {
    while (true) {
        const state = ns.go.resetBoardState(opponent, boardSize).map(row => row.replaceAll("#", "").split(''));

        if (isRect(state, boardSize)) {
            return { boardX: state.length, boardY: state[0].length };
        }

        await ns.sleep(10);
    }
}

/**
 * @param {string[][]} matrix
 * @param {number} boardSize
 */
const isRect = (matrix, boardSize) => {
    if (!Array.isArray(matrix) || matrix.length !== boardSize || !Array.isArray(matrix[0])) {
        return false;
    }
    const numCols = matrix[0].length;
    for (let i = 1; i < matrix.length; i++) {
        if (!Array.isArray(matrix[i]) || matrix[i].length !== numCols) {
            return false;
        }
    }
    return true;
}

const lettersToNumbers = {
    A: 0, B: 1, C: 2, D: 3, E: 4, F: 5, G: 6, H: 7, J: 8, K: 9, L: 10, M: 11, N: 12
}

const numbersToLetters = {
    0: "A", 1: "B", 2: "C", 3: "D", 4: "E", 5: "F", 6: "G", 7: "H", 8: "J", 9: "K", 10: "L", 11: "M", 12: "N"
};
