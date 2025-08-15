import {
  K8sNameDescriptionFieldData,
  K8sNameDescriptionFieldUpdateFunction,
  K8sNameDescriptionType,
} from '@odh-dashboard/internal/concepts/k8s/K8sNameDescriptionField/types';
import { useModelTypeField, type ModelTypeFieldData } from './fields/ModelTypeSelectField';
import { useDeploymentName } from './fields/ModelNameInputField';

export type UseModelDeploymentWizardProps = {
  modelTypeField?: ModelTypeFieldData;
  setModelType?: (data: ModelTypeFieldData) => void;
  deploymentName?: K8sNameDescriptionFieldData | K8sNameDescriptionType;
  setDeploymentName?: K8sNameDescriptionFieldUpdateFunction;
  // Add more field handlers as needed
};
export const useModelDeploymentWizard = (
  existingData?: UseModelDeploymentWizardProps,
): UseModelDeploymentWizardProps => {
  const [modelType, setModelType] = useModelTypeField(existingData?.modelTypeField);
  const { data: deploymentName, onDataChange: setDeploymentName } = useDeploymentName(
    existingData?.deploymentName,
  );

  return {
    modelTypeField: modelType,
    setModelType,
    deploymentName,
    setDeploymentName,
  };
};
