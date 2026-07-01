import React, { useState } from "react";
import {
  ActionGroup,
  Alert,
  Button,
  Form,
  FormGroup,
  TextInput,
  LoginPage as PFLoginPage,
} from "@patternfly/react-core";

import { brandingStrings } from "./branding";
import type { LoginConfig } from "./types";

interface DeviceVerifyPageProps {
  config: LoginConfig;
}

// DeviceVerifyPage renders the device-code entry form using PatternFly 6
// components. The form action is provided by the hub-injected config;
// branding and page text come from the build-time bundle.
export const DeviceVerifyPage: React.FC<DeviceVerifyPageProps> = ({ config }) => {
  const [userCode, setUserCode] = useState("");

  const { deviceFormAction } = config;
  const { devicePage, styles, application } = brandingStrings;

  return (
    <PFLoginPage
      brandImgSrc={styles?.brandImage ?? ""}
      brandImgAlt={application?.name ?? application?.title ?? ""}
      backgroundImgSrc={styles?.backgroundImage ?? ""}
      loginTitle={devicePage?.title ?? "Device Authorization"}
      loginSubtitle={
        devicePage?.subtitle ?? "Enter the code displayed on your device"
      }
    >
      <Alert
        variant="info"
        isInline
        isPlain
        title="Enter the user code shown on your device to grant access."
        style={{ marginBottom: "var(--pf-t--global--spacer--md)" }}
      />
      <form method="post" action={deviceFormAction ?? ""}>
        <Form>
          <FormGroup label="User Code" fieldId="userCode">
            <TextInput
              id="userCode"
              name="userCode"
              type="text"
              value={userCode}
              onChange={(_e, v) => setUserCode(v.toUpperCase())}
              placeholder="XXXX-XXXX"
              autoFocus
              required
            />
          </FormGroup>
          <ActionGroup>
            <Button
              variant="primary"
              type="submit"
              isDisabled={userCode.trim().length === 0}
            >
              Authorize Device
            </Button>
          </ActionGroup>
        </Form>
      </form>
    </PFLoginPage>
  );
};
