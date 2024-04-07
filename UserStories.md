# User stories defining main functionality

*As a website user, I would like to* 

see the parts for all year versions of specific model in one view

*so that*

I would not need to view each year individually.

### DoD:
- Parts for a specific model of all available years can be displayed
- Some year information for each part should still be present

---

*As a website user, I would like to* 

see only the brands that match my selected vehicle type

*so that*

I would not have to pay attention to unrelevant brands

### DoD:
- Functionality to choose vehicle type is available
- After selecting a vehicle type, vehicles/parts of other vehicle types are not displayed.
---

*As a website user, I would like to* 

see which parts of other models are compatible with a specific vehicle model + year

*so that*

I could find the needed part more easily

### DoD:
- Compatibility between different models parts is shown
- A view is available, that displays parts of one or more models that fit to a specific vehicle's model.
---

*As a website user, I would like to*

find the correct vehicle model easily by using search functionality

*so that*

I would not need to spend time narrowing the filters if I know the model name right away.

### DoD:
- A search bar is available.
- Search bar includes basic Boolean operators and free text search.
---

*As a website user, I would like to*

see what vehicles are new and when they were added

*so that*

I would know if I've already seen the parts of a vehicle

### DoD:
- Timestamp of last update for a vehicle or it's parts is visible to the user
- A label is added to a vehicle if its parts have been updated, or if the vehicle itself is new.
---

*As a website user, I would like to*

order vehicles based on the timestamp of last update

*so that*

I could find recently updated vehicles easily

### DoD:
- A button exists that allows me to sort the vehicles based on update time
---

*As a website user, I would like to*

order parts based on the timestamp of last update

*so that*

I could see which parts have been added most recently

### DoD:
- A button exist that sorts the parts based on time
- Timestamp of last update is returned with API response.
---