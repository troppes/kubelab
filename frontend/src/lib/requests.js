import { error } from '@sveltejs/kit';

export async function get(token, url) {
    const response = await fetch(url, {
        method: 'GET',
        headers: {
            'Content-type': 'application/json',
            'Authorization': `${token}`,
        },
    })
    const json = await response.json();
    if (response.ok) {
        return json;
    } else {
        throw new error(response.status, json);
    }
}

export async function post(token, data, url) {
    const response = await fetch(url, {
        method: 'POST',
        headers: {
            'Content-type': 'application/json',
            'Authorization': `${token}`,
        },
        body: JSON.stringify(data),
    })
    const json = await response.json();
    if (response.ok) {
        return json;
    } else {
        throw new error(response.status, json);
    }
}

export async function put(token, data, url) {
    const response = await fetch(url, {
        method: 'PUT',
        headers: {
            'Content-type': 'application/json',
            'Authorization': `${token}`,
        },
        body: JSON.stringify(data),
    })
    const json = await response.json();
    if (response.ok) {
        return json;
    } else {
        throw new error(response.status, json);
    }
}

export async function deleteReq(token, url) { // delete is a special word
    const response = await fetch(url, {
        method: 'DELETE',
        headers: {
            'Content-type': 'application/json',
            'Authorization': `${token}`,
        },
    })
    const json = await response.json();
    if (response.ok) {
        return json;
    } else {
        throw new error(response.status, json);
    }
}