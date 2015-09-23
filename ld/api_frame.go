package ld

// EmbedNode represents embed meta info
type EmbedNode struct {
	parent   interface{}
	property string
}

// FramingContext stores framing state
type FramingContext struct {
	embed       bool
	explicit    bool
	omitDefault bool
	embeds      map[string]*EmbedNode
}

// NewFramingContext creates and returns as new framing context.
func NewFramingContext(opts *JsonLdOptions) *FramingContext {
	context := &FramingContext{
		embed:       true,
		explicit:    false,
		omitDefault: false,
	}

	if opts != nil {
		context.embed = opts.Embed
		context.explicit = opts.Explicit
		context.omitDefault = opts.OmitDefault
	}

	return context
}

// Frame performs JSON-LD framing as defined in:
//
// http://json-ld.org/spec/latest/json-ld-framing/
//
// Frames the given input using the frame according to the steps in the Framing Algorithm.
// The input is used to build the framed output and is returned if there are no errors.
//
// Returns the framed output.
func (api *JsonLdApi) Frame(input interface{}, frame []interface{}, opts *JsonLdOptions) ([]interface{}, error) {
	idGen := NewBlankNodeIDGenerator()

	// create framing state
	state := NewFramingContext(opts)

	nodes := make(map[string]interface{})
	api.GenerateNodeMap(input, nodes, "@default", nil, "", nil, idGen)
	nodeMap := nodes["@default"].(map[string]interface{})

	framed := make([]interface{}, 0)

	// NOTE: frame validation is done by the function not allowing anything
	// other than list to be passed
	var frameParam map[string]interface{}
	if frame != nil && len(frame) > 0 {
		frameParam = frame[0].(map[string]interface{})
	} else {
		frameParam = make(map[string]interface{})
	}
	framedObj, _ := api.frame(state, nodeMap, nodeMap, frameParam, framed, "")
	// because we know framed is an array, we can safely cast framedObj back to an array
	return framedObj.([]interface{}), nil
}

// frame subjects according to the given frame.
// state: the current framing state
// nodes:
// nodeMap: node map
// frame: the frame
// parent: the parent subject or top-level array
// property: the parent property, initialized to nil
func (api *JsonLdApi) frame(state *FramingContext, nodes map[string]interface{}, nodeMap map[string]interface{},
	frame map[string]interface{}, parent interface{}, property string) (interface{}, error) {

	// filter out subjects that match the frame
	matches, err := FilterNodes(nodes, frame)
	if err != nil {
		return nil, err
	}

	// get flags for current frame
	embedOn := GetFrameFlag(frame, "@embed", state.embed)
	explicitOn := GetFrameFlag(frame, "@explicit", state.explicit)

	// add matches to output
	for _, id := range GetOrderedKeys(matches) {
		if property == "" {
			state.embeds = make(map[string]*EmbedNode)
		}

		// start output
		output := make(map[string]interface{})
		output["@id"] = id

		// prepare embed meta info
		embeddedNode := &EmbedNode{}
		embeddedNode.parent = parent
		embeddedNode.property = property

		// if embed is on and there is an existing embed
		if existing, hasID := state.embeds[id]; embedOn && hasID {
			embedOn = false

			if parentList, isList := existing.parent.([]interface{}); isList {
				for _, p := range parentList {
					if CompareValues(output, p) {
						embedOn = true
						break
					}
				}
			} else {
				// existing embed's parent is an object
				parentMap := existing.parent.(map[string]interface{})
				if propertyVal, hasProperty := parentMap[existing.property]; hasProperty {
					for _, v := range propertyVal.([]interface{}) {
						if vMap, isMap := v.(map[string]interface{}); isMap && vMap["@id"] == id {
							embedOn = true
							break
						}
					}
				}
			}

			// existing embed has already been added, so allow an overwrite
			if embedOn {
				removeEmbed(state, id)
			}
		}

		// not embedding, add output without any other properties
		if !embedOn {
			parent = addFrameOutput(parent, property, output)
		} else {
			// add embed meta info
			state.embeds[id] = embeddedNode

			// iterate over subject properties
			element := matches[id].(map[string]interface{})
			for _, prop := range GetOrderedKeys(element) {

				// copy keywords to output
				if IsKeyword(prop) {
					output[prop] = CloneDocument(element[prop])
					continue
				}

				// if property isn't in the frame
				if _, containsProp := frame[prop]; !containsProp {
					// if explicit is off, embed values
					if !explicitOn {
						api.embedValues(state, nodeMap, element, prop, output)
					}
					continue
				}

				// add objects
				value := element[prop].([]interface{})

				for _, item := range value {
					// recurse into list
					itemMap, isMap := item.(map[string]interface{})
					listValue, hasList := itemMap["@list"]
					if isMap && hasList {
						// add empty list
						list := make(map[string]interface{})
						list["@list"] = make([]interface{}, 0)
						addFrameOutput(output, prop, list)

						// add list objects
						for _, listitem := range listValue.([]interface{}) {
							// recurse into subject reference
							if IsNodeReference(listitem) {
								tmp := make(map[string]interface{})
								itemid := listitem.(map[string]interface{})["@id"].(string)
								// TODO: nodes may need to be node_map,
								// which is global
								tmp[itemid] = nodeMap[itemid]
								api.frame(state, tmp, nodeMap, frame[prop].([]interface{})[0].(map[string]interface{}),
									list, "@list")
							} else {
								// include other values automatcially (TODO:
								// may need Clone(n)
								addFrameOutput(list, "@list", listitem)
							}
						}
					} else if IsNodeReference(item) {
						// recurse into subject reference
						tmp := make(map[string]interface{})
						itemid := itemMap["@id"].(string)
						// TODO: nodes may need to be node_map, which is
						// global
						tmp[itemid] = nodeMap[itemid]
						api.frame(state, tmp, nodeMap, frame[prop].([]interface{})[0].(map[string]interface{}), output,
							prop)
					} else {
						// include other values automatically (TODO: may
						// need Clone(o)
						addFrameOutput(output, prop, item)
					}
				}
			}

			// handle defaults
			for _, prop := range GetOrderedKeys(frame) {
				// skip keywords
				if IsKeyword(prop) {
					continue
				}

				pf := frame[prop].([]interface{})
				var propertyFrame map[string]interface{}
				if len(pf) > 0 {
					propertyFrame = pf[0].(map[string]interface{})
				}

				if propertyFrame == nil {
					propertyFrame = make(map[string]interface{})
				}

				omitDefaultOn := GetFrameFlag(propertyFrame, "@omitDefault", state.omitDefault)
				if _, hasProp := output[prop]; !omitDefaultOn && !hasProp {
					var def interface{} = "@null"
					if defaultVal, hasDefault := propertyFrame["@default"]; hasDefault {
						def = CloneDocument(defaultVal)
					}
					if _, isList := def.([]interface{}); !isList {
						def = []interface{}{def}
					}
					output[prop] = []interface{}{
						map[string]interface{}{
							"@preserve": def,
						},
					}
				}
			}

			// add output to parent
			parent = addFrameOutput(parent, property, output)
		}
	}
	return parent, nil
}

// GetFrameFlag gets the frame flag value for the given flag name.
// If boolean value is not found, returns theDefault
func GetFrameFlag(frame map[string]interface{}, name string, theDefault bool) bool {
	value := frame[name]
	switch v := value.(type) {
	case []interface{}:
		if len(v) > 0 {
			value = v[0]
		}
	case map[string]interface{}:
		if valueVal, present := v["@value"]; present {
			value = valueVal
		}
	case bool:
		return v
	}

	if valueBool, isBool := value.(bool); isBool {
		return valueBool
	}

	return theDefault
}

// removeEmbed removes an existing embed with the given id.
func removeEmbed(state *FramingContext, id string) {
	// get existing embed
	embeds := state.embeds
	embed := embeds[id]
	parent := embed.parent
	property := embed.property

	// create reference to replace embed
	node := make(map[string]interface{})
	node["@id"] = id

	// remove existing embed
	if IsNode(parent) {
		// replace subject with reference
		newVals := make([]interface{}, 0)
		parentMap := parent.(map[string]interface{})
		oldvals := parentMap[property].([]interface{})
		for _, v := range oldvals {
			vMap, isMap := v.(map[string]interface{})
			if isMap && vMap["@id"] == id {
				newVals = append(newVals, node)
			} else {
				newVals = append(newVals, v)
			}
		}
		parentMap[property] = newVals
	}
	// recursively remove dependent dangling embeds
	removeDependents(embeds, id)
}

// removeDependents recursively removes dependent dangling embeds.
func removeDependents(embeds map[string]*EmbedNode, id string) {
	// get embed keys as a separate array to enable deleting keys in map
	for idDep, e := range embeds {
		var p map[string]interface{}
		if e.parent != nil {
			var isMap bool
			p, isMap = e.parent.(map[string]interface{})
			if !isMap {
				continue
			}
		} else {
			p = make(map[string]interface{})
		}

		pid := p["@id"].(string)
		if id == pid {
			delete(embeds, idDep)
			removeDependents(embeds, idDep)
		}
	}
}

// FilterNodes returns a map of all of the nodes that match a parsed frame.
func FilterNodes(nodes map[string]interface{}, frame map[string]interface{}) (map[string]interface{}, error) {
	rval := make(map[string]interface{})
	for id, elementVal := range nodes {
		element, _ := elementVal.(map[string]interface{})
		if element != nil {
			if res, err := FilterNode(element, frame); res {
				if err != nil {
					return nil, err
				}
				rval[id] = element
			}
		}
	}
	return rval, nil
}

// FilterNode returns true if the given node matches the given frame.
func FilterNode(node map[string]interface{}, frame map[string]interface{}) (bool, error) {
	types, _ := frame["@type"]
	if types != nil {
		typesList, isList := types.([]interface{})
		if !isList {
			return false, NewJsonLdError(SyntaxError, "frame @type must be an array")
		}
		nodeTypesVal, nodeHasType := node["@type"]
		var nodeTypes []interface{}
		if !nodeHasType {
			nodeTypes = make([]interface{}, 0)
		} else if nodeTypes, isList = nodeTypesVal.([]interface{}); !isList {
			return false, NewJsonLdError(SyntaxError, "node @type must be an array")
		}
		if len(typesList) == 1 {
			vMap, isMap := typesList[0].(map[string]interface{})
			if isMap && len(vMap) == 0 {
				return len(nodeTypes) > 0, nil
			}
		}

		for _, i := range nodeTypes {
			for _, j := range typesList {
				if DeepCompare(i, j, false) {
					return true, nil
				}
			}
		}
		return false, nil
	}

	for _, key := range GetKeys(frame) {
		_, nodeContainsKey := node[key]
		if key == "@id" || !IsKeyword(key) && !nodeContainsKey {
			return false, nil
		}
	}
	return true, nil
}

// addFrameOutput adds framing output to the given parent.
// parent: the parent to add to.
// property: the parent property.
// output: the output to add.
func addFrameOutput(parent interface{}, property string, output interface{}) interface{} {
	if parentMap, isMap := parent.(map[string]interface{}); isMap {
		propVal, hasProperty := parentMap[property]
		if hasProperty {
			parentMap[property] = append(propVal.([]interface{}), output)
		} else {
			parentMap[property] = []interface{}{output}
		}
		return parentMap
	}

	return append(parent.([]interface{}), output)
}

// embedValues embeds values for the given subject [element] and property into the given output
// during the framing algorithm.
func (api *JsonLdApi) embedValues(state *FramingContext, nodeMap map[string]interface{},
	element map[string]interface{}, property string, output interface{}) {
	// embed subject properties in output
	objects := element[property].([]interface{})
	for _, o := range objects {
		oMap, isMap := o.(map[string]interface{})
		_, hasList := oMap["@list"]
		if isMap && hasList {
			list := make(map[string]interface{})
			list["@list"] = make([]interface{}, 0)
			api.embedValues(state, nodeMap, oMap, "@list", list)
			addFrameOutput(output, property, list)
		} else if IsNodeReference(o) {
			// handle subject reference
			oMap := o.(map[string]interface{})
			sid := oMap["@id"].(string)

			// embed full subject if isn't already embedded
			if _, hasSID := state.embeds[sid]; !hasSID {
				// add embed
				state.embeds[sid] = &EmbedNode{
					parent:   output,
					property: property,
				}

				// recurse into subject
				o = make(map[string]interface{})
				s, hasSID := nodeMap[sid]
				sMap, isMap := s.(map[string]interface{})
				if !hasSID || !isMap {
					sMap = map[string]interface{}{
						"@id": sid,
					}
				}
				for prop, propValue := range sMap {
					// copy keywords
					if IsKeyword(prop) {
						o.(map[string]interface{})[prop] = CloneDocument(propValue)
						continue
					}
					api.embedValues(state, nodeMap, sMap, prop, o)
				}
			}
			addFrameOutput(output, property, o)
		} else {
			// copy non-subject value
			addFrameOutput(output, property, CloneDocument(o))
		}
	}
}
