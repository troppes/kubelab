import { SvelteKitAuth } from "@auth/sveltekit";
import Keycloak from "@auth/core/providers/keycloak";
import { env } from '$env/dynamic/private';
import { decode } from '$lib/helpers.js';

// needed since env variables are a string
if (env.ALLOW_UNTRUSTED_CERTS === "true" || env.ALLOW_UNTRUSTED_CERTS === "TRUE") {
  process.env["NODE_TLS_REJECT_UNAUTHORIZED"] = 0;
}

export const handle = SvelteKitAuth({
  providers: [
    Keycloak({
      clientId: env.KEYCLOAK_CLIENT_ID,
      clientSecret: env.KEYCLOAK_CLIENT_SECRET,
      issuer: env.KEYCLOAK_ISSUER_URL,
    }),
  ],
  callbacks: {
    async jwt({ token, user, account, profile, isNewUser }) {
      if (account) {
        const token = decode(account.access_token)
        let newToken = {
          "name": token.name,
          "email": token.email,
          "given_name": token.given_name,
          "family_name": token.family_name,
          "id_token": account.id_token,
          "roles": token.groups, // This is possible thanks to the mapping done before for Kubernetes to identify the roles
          "user_id": token.user_id,
          "username": token.preferred_username
        }
        return newToken;
      }
      return token;
    },
    async session({ session, user, token }) {
      session.user = token;
      return session;
    },
  },
});

