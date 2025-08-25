import React from 'react';
import { z } from 'zod';
import {
  Form,
  FormSection,
  useWizardContext,
  useWizardFooter,
  WizardFooter,
} from '@patternfly/react-core';
import { UseModelDeploymentWizardProps } from '../useDeploymentWizard';
import ProjectSection from '../fields/ProjectSection';
import {
  deploymentNameInputFieldSchema,
  isK8sNameDescriptionFieldData,
  isK8sNameDescriptionTypeOnly,
  ModelNameInputField,
  validateDeploymentNameInput,
} from '../fields/ModelNameInputField';

const modelDeploymentStepSchema = z.object({
  deploymentName: deploymentNameInputFieldSchema,
});

export type ModelDeploymentStepData = z.infer<typeof modelDeploymentStepSchema>;

type ModelDeploymentStepProps = {
  projectName: string;
  wizardData: UseModelDeploymentWizardProps;
};

export const ModelDeploymentStepContent: React.FC<ModelDeploymentStepProps> = ({
  projectName,
  wizardData,
}) => {
  const { activeStep, goToNextStep, goToPrevStep, close } = useWizardContext();

  const [showNonEmptyNameWarning, setShowNonEmptyNameWarning] = React.useState(false);

  useWizardFooter(
    <WizardFooter
      activeStep={activeStep}
      onNext={() => {
        if (
          isK8sNameDescriptionFieldData(wizardData.deploymentName) &&
          validateDeploymentNameInput({
            name: wizardData.deploymentName.name,
            k8sName: wizardData.deploymentName.k8sName.value,
          }).isValid
        ) {
          setShowNonEmptyNameWarning(false);
          goToNextStep();
        }
        setShowNonEmptyNameWarning(true);
      }}
      onBack={goToPrevStep}
      isBackDisabled={activeStep.index === 1}
      onClose={close}
      nextButtonText="Next"
    />,
  );

  return (
    <Form>
      <FormSection title="Model deployment">
        {projectName && <ProjectSection projectName={projectName} />}
        {isK8sNameDescriptionFieldData(wizardData.deploymentName) && (
          <ModelNameInputField
            deploymentName={wizardData.deploymentName}
            setDeploymentName={wizardData.setDeploymentName}
            showNonEmptyNameWarning={showNonEmptyNameWarning}
          />
        )}
      </FormSection>
    </Form>
  );
};
