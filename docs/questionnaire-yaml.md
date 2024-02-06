
## Questionnaire Documentation

Here, we will describe the fields and their purpose to make creating a new assessment easy

### Questionnaire Fields

1. name:  Name of the questionnaire. Name is required and must be unique for the entire konveyor instance.
2. description: An optional short description of the questionnaire.
3. thresholds: Required the threshold definition for each risk category for the application or the archetype to be considered affected by that risk level. The higher risk level always takes precedence. For example, if yellow threshold is established in 30% and red in 5%, and the answers for an application or archetype have 35% yellow and 6% red, the risk level for the application or archetype will be red.
    1. red: Required numeric percentage (example: 30 for 30%) of red answers the questionnaire can have until the risk level is considered red.
    2. yellow: Required numeric percentage (example: 30 for 30%) of yellow answers the questionnaire can have until the risk level is considered yellow.
    3. unknown: Required numeric percentage (example: 30 for 30%) of unknown answers the questionnaire can have until the risk level is considered unknown.
4. riskMessages: Required messages to be displayed in reports for each risk category. The risk_messages map is defined by the following fields:
    5. red: Required string message to display in reports for the red risk level.
    6. yellow: Required string message to display in reports for the yellow risk level.
    7. green: Required string message to display in reports for the green risk level.
    8. unknown: Required string message to display in reports for the unknown risk level.
5. sections: Required list of sections that the questionnaire will include.
    1. name: Name is the required string to be displayed for the section.
    1. order: Required int order in which the question should appear in the section.
    1. comment: Optional string to describe the section.
    1. questions: Required list of questions that belong to the section. 
        1. order: Required int order in which the question should appear in the section.
        1. text: Required string of the question to be asked.
        1. explanation: Optional string of additional explanations for the question.
        2. includeFor: Optional list that defines a question should be displayed if any of the tags included in the list is present in the target application or archetype.
            1. category: Required string category of the target tag.
            2. tag: Required string for the target tag.
        3. excludeFor: Optional list defines that a question should be skipped if any of the tags included in the list is present in the target application or archetype.
            1. category: Required string category of the target tag.
            2. tag: Required string for the target tag.
        4. answers: Required list of answers for the given question. 
            1. order: Required int order in which the question should appear in the section.
            1. text:  Required string the actual answer for the question.
            2. risk: Required to be one of red, yellow, green, or unknown. The risk level the current answer implies.
            3. rationale:  Optional string explaining the justification for the answer being considered a risk.
            4. mitigation: Optional string for an explanation of the potential mitigation strategy for the risk implied by this answer.
            5. applyTags: Optional list that defines a list of tags to be automatically applied to the assessed application or archetype if this answer is selected.
                1. category: Required string category of the target tag.
                2. tag: Required string for the target tag.
            6. autoAnswerFor: Optional list defines a list of tags that will lead to this answer being automatically selected when the application or archetype is assessed. 
                1. category: Required string category of the target tag.
                2. tag: Required string for the target tag.

> [!NOTE]
> 1. Anything with the word **required** must be filed out. Otherwise, the yaml will not validate on upload.
> 2. Each subsection defines a new struct/object in yaml. For instance

> ```yaml
> ...
> name: Testing
> thresholds: 
>     red: 30
>     yellow: 45
>     unknown: 5
> ...
> ```


