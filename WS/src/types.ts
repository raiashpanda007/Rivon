const InvalidJSON = {
  type: "ERROR",
  payload: {
    message: "Invalid JSON . Unable to parse JSON"
  }
}
const Types = {
  ErrorMessage: {
    InvalidJSON: InvalidJSON
  }
}
export default Types;
