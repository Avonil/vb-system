// Типи для відповідей Твіча
interface TwitchTokenResponse {
    access_token: string;
    refresh_token: string;
    expires_in: number;
    scope: string[];
}

interface TwitchUser {
    id: string;
    login: string;
    display_name: string;
    profile_image_url: string;
    email?: string;
}

export const getTwitchToken = async (code: string) => {
    const params = new URLSearchParams({
        client_id: process.env.TWITCH_CLIENT_ID!,
        client_secret: process.env.TWITCH_CLIENT_SECRET!,
        code,
        grant_type: 'authorization_code',
        redirect_uri: process.env.TWITCH_REDIRECT_URI! || "https://185.183.92.97:3000/auth/callback"
    });

    const response = await fetch('https://id.twitch.tv/oauth2/token', {
        method: 'POST',
        body: params
    });

    if (!response.ok) throw new Error('Failed to fetch Twitch token');
    return await response.json() as TwitchTokenResponse;
};

export const getTwitchUser = async (accessToken: string) => {
    const response = await fetch('https://api.twitch.tv/helix/users', {
        headers: {
            'Client-ID': process.env.TWITCH_CLIENT_ID!,
            'Authorization': `Bearer ${accessToken}`
        }
    });

    if (!response.ok) throw new Error('Failed to fetch user data');
    const data = await response.json() as { data: TwitchUser[] };
    return data.data[0];
};