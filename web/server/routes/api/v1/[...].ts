import { createBffHandler } from "@yueli/nuxt-runtime/server";
import { withYtvBffErrors } from "../../../utils/ytv-bff";

// Opening the review editor establishes a visitor session before its first write.
const handler = createBffHandler({
  mountPath: "/api/v1",
  resolveTarget({ event }) {
    return identityBffTarget(oidcConfig(event).downstreamBase);
  },
  credential: {
    async resolve({ event }) {
      const personalToken = identityPersonalTokenCredential(getHeader(event, "authorization"));
      if (personalToken) return personalToken;
      let headers = await sessionAuthHeaders(event);
      const isReview = /^\/api\/v1\/me\/source-reviews\/[^/]+$/.test(getRequestURL(event).pathname);
      if (!headers.authorization && isReview) {
        headers = await guestSessionAuthHeaders(event, String(useRuntimeConfig(event).ytvAudience), isReview);
      }
      return identityBffCredential(headers);
    },
  },
});

export default withYtvBffErrors(handler);
