import React from "react";
import {
  EmptyState,
  EmptyStateBody,
  EmptyStateVariant,
  LoginPage as PFLoginPage,
} from "@patternfly/react-core";
import { ExclamationCircleIcon } from "@patternfly/react-icons";

import { brandingStrings } from "./branding";
import type { LoginConfig } from "./types";

interface SessionExpiredPageProps {
  config: LoginConfig;
}

// SessionExpiredPage is shown when the login session has completely expired
// and cannot be renewed. No form is presented; the user must return to the
// application to restart the authentication flow.
export const SessionExpiredPage: React.FC<SessionExpiredPageProps> = (
  _props: SessionExpiredPageProps
) => {
  const { styles, application } = brandingStrings;

  return (
    <PFLoginPage
      loginTitle="Session Expired"
      backgroundImgSrc={styles?.backgroundImage ?? ""}
      brandImgProps={{
        src: styles?.brandImage ?? "",
        alt: application?.name ?? application?.title ?? "",
        widths: { default: "260px" },
      }}
      textContent=""
    >
      <EmptyState
        variant={EmptyStateVariant.lg}
        icon={ExclamationCircleIcon}
        status="danger"
        titleText="Session Expired"
        headingLevel="h2"
      >
        <EmptyStateBody>
          Your login session has expired. Please return to the application and
          try again.
        </EmptyStateBody>
      </EmptyState>
    </PFLoginPage>
  );
};
