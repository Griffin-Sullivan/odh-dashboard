import K8sNameDescriptionField, {
  useK8sNameDescriptionFieldData,
} from '@odh-dashboard/internal/concepts/k8s/K8sNameDescriptionField/K8sNameDescriptionField';
import {
  K8sNameDescriptionFieldData,
  K8sNameDescriptionType,
  UseK8sNameDescriptionFieldData,
  K8sNameDescriptionFieldUpdateFunction,
} from '@odh-dashboard/internal/concepts/k8s/K8sNameDescriptionField/types.js';
import React from 'react';
import { FormGroup } from '@patternfly/react-core';
import { isK8sNameDescriptionType } from '@odh-dashboard/internal/concepts/k8s/K8sNameDescriptionField/utils.js';
import { z } from 'zod';

// Schema
export const deploymentNameInputFieldSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, 'Name cannot be empty')
    .refine((val) => val.trim().length > 0, 'Name cannot be empty'),
  k8sName: z
    .string()
    .min(1, 'Resource name cannot be empty')
    .max(253, 'Resource name cannot exceed 253 characters')
    .regex(
      /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/,
      'Must consist of lowercase alphanumeric characters or "-", and must start and end with an alphanumeric character',
    ),
});

export type DeploymentNameInputData = z.infer<typeof deploymentNameInputFieldSchema>;

// Helper function to validate data using the schema
export const validateDeploymentNameInput = (
  data: DeploymentNameInputData,
): { isValid: boolean; errors?: z.ZodError } => {
  const result = deploymentNameInputFieldSchema.safeParse(data);
  if (result.success) {
    return { isValid: true };
  }
  return { isValid: false, errors: result.error };
};

// Custom type guard that only accepts K8sNameDescriptionType
export const isK8sNameDescriptionTypeOnly = (
  x?: K8sNameDescriptionType | K8sNameDescriptionFieldData,
): x is K8sNameDescriptionType => {
  return !!x && 'k8sName' in x && typeof x.k8sName === 'string';
};

// Helper function to check K8sNameDescriptionFieldData type
export const isK8sNameDescriptionFieldData = (
  data?: K8sNameDescriptionFieldData | K8sNameDescriptionType,
): data is K8sNameDescriptionFieldData => {
  return !!data && !isK8sNameDescriptionTypeOnly(data);
};

export const useDeploymentName = (
  existingData?: K8sNameDescriptionFieldData | K8sNameDescriptionType,
): UseK8sNameDescriptionFieldData => {
  if (isK8sNameDescriptionTypeOnly(existingData)) {
    return useK8sNameDescriptionFieldData({ initialData: existingData });
  }
  return useK8sNameDescriptionFieldData();
};

// Component

type ModelNameInputFieldProps = {
  deploymentName: K8sNameDescriptionFieldData;
  setDeploymentName?: K8sNameDescriptionFieldUpdateFunction;
  showNonEmptyNameWarning?: boolean;
};

export const ModelNameInputField: React.FC<ModelNameInputFieldProps> = ({
  deploymentName,
  setDeploymentName,
  showNonEmptyNameWarning,
}) => {
  return (
    <FormGroup fieldId="model-name-input" isRequired>
      <K8sNameDescriptionField
        data={deploymentName}
        onDataChange={setDeploymentName}
        dataTestId="model-name"
        nameLabel="Model name"
        nameHelperText="This is the name of the inference service created when the model is deployed." // TODO: fix to be not KServe specific
        hideDescription
        showNonEmptyNameWarning={showNonEmptyNameWarning && deploymentName.name.trim().length === 0}
      />
    </FormGroup>
  );
};
