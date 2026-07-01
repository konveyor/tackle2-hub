import React from "react";
import {
  EmptyState,
  EmptyStateBody,
  EmptyStateVariant,
  LoginPage as PFLoginPage,
} from "@patternfly/react-core";
import { CheckCircleIcon } from "@patternfly/react-icons";

import { brandingStrings } from "./branding";
import type { LoginConfig } from "./types";

interface DeviceSuccessPageProps {
  config: LoginConfig;
}

// DeviceSuccessPage renders the confirmation shown after successful device
// authorization. No dynamic data is needed -- all text comes from build-time
// branding strings.
export const DeviceSuccessPage: React.FC<DeviceSuccessPageProps> = (_props: DeviceSuccessPageProps) => {
  const { devicePage, styles, application } = brandingStrings;

  return (
    <PFLoginPage
      brandImgSrc={styles?.brandImage ?? ""}
      brandImgAlt={application?.name ?? application?.title ?? ""}
      backgroundImgSrc={styles?.backgroundImage ?? ""}
      loginTitle={devicePage?.successTitle ?? "Authorization Complete"}
    >
      <EmptyState
        variant={EmptyStateVariant.lg}
        icon={CheckCircleIcon}
        status="success"
        titleText={devicePage?.successTitle ?? "Authorization Complete"}
        headingLevel="h2"
      >
        <EmptyStateBody>
          {devicePage?.successMessage ??
            "You have successfully authorized the device. You may close this window."}
        </EmptyStateBody>
      </EmptyState>
    </PFLoginPage>
  );
};
