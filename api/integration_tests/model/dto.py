import json
from typing import Dict, Union, Optional

class CreateFolderDTO():
    name: str
    parent_id: Optional[int]

    def __init__(self, name: str, parent_id: Optional[int]) -> None:
        self.id = id
        self.name = name
        self.parent_id = parent_id

    def toJson(self) -> str:
        obj: Dict[str, Union[str, int]] = dict()
        obj['name'] = self.name
        if self.parent_id is not None:
            obj['parentId'] = self.parent_id
        return json.dumps(obj)

class UpdateFolderDTO(CreateFolderDTO):
    pass
