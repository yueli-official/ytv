import handler from "@yueli/identity-nuxt/server/routes/api/v1/[...].ts";
import { withYtvBffErrors } from "../../../utils/ytv-bff";

export default withYtvBffErrors(handler);
